package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
	"unicode"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sirupsen/logrus"
	mint "github.com/void616/gm.mint"
	"github.com/void616/gm.mint.rpc/request"
	"github.com/void616/gm.mint.sender/internal/alert"
	"github.com/void616/gm.mint.sender/internal/metrics"
	"github.com/void616/gm.mint.sender/internal/mint/blockobserver"
	"github.com/void616/gm.mint.sender/internal/mint/blockparser"
	"github.com/void616/gm.mint.sender/internal/mint/blockranger"
	"github.com/void616/gm.mint.sender/internal/mint/rpcpool"
	"github.com/void616/gm.mint.sender/internal/mint/txfilter"
	serviceAPI "github.com/void616/gm.mint.sender/internal/sender/api"
	serviceHTTP "github.com/void616/gm.mint.sender/internal/sender/api/http"
	serviceNats "github.com/void616/gm.mint.sender/internal/sender/api/nats"
	"github.com/void616/gm.mint.sender/internal/sender/db"
	"github.com/void616/gm.mint.sender/internal/sender/db/mysql"
	"github.com/void616/gm.mint.sender/internal/sender/db/types"
	"github.com/void616/gm.mint.sender/internal/sender/notifier"
	"github.com/void616/gm.mint.sender/internal/sender/txconfirmer"
	"github.com/void616/gm.mint.sender/internal/sender/txsigner"
	"github.com/void616/gm.mint.sender/internal/version"
	"github.com/void616/gm.mint/signer"
	"github.com/void616/gm.mint/transaction"
	"github.com/void616/gotask"
	"gopkg.in/yaml.v2"
)

var (
	logger            *logrus.Logger
	group             *gotask.Group
	blockObserverTask *gotask.Task
	blockRangerTask   *gotask.Task
	txFilterTask      *gotask.Task
	txConfirmerTask   *gotask.Task
	natsTransportTask *gotask.Task
	httpTransportTask *gotask.Task
	notifierTask      *gotask.Task
	txSignerTask      *gotask.Task
	metricsTask       *gotask.Task
)

func main() {
	var argConfigFile = flag.String("config", "./config.yaml", "Config Yaml")
	flag.Parse()

	// config file
	var conf config
	{
		b, err := ioutil.ReadFile(*argConfigFile)
		if err != nil {
			panic("failed to read config file: " + err.Error())
		}
		if err := yaml.Unmarshal(b, &conf); err != nil {
			panic("failed to parse config file: " + err.Error())
		}
	}

	// logging
	logger = logrus.New()
	var tasksLogger gotask.Logger
	{
		lvl, err := logrus.ParseLevel(conf.Log.Level)
		if err != nil {
			logger.WithError(err).Fatal("Failed to parse logger level")
		}
		logger.Out = os.Stdout
		logger.Level = lvl

		// format
		switch {
		case conf.Log.JSON:
			logger.Formatter = &logrus.JSONFormatter{
				TimestampFormat: time.RFC3339,
			}
		default:
			logger.Formatter = &logrus.TextFormatter{
				FullTimestamp:   true,
				DisableColors:   !conf.Log.Color,
				ForceColors:     conf.Log.Color,
				TimestampFormat: time.RFC3339,
			}
		}

		// trace level tweaks
		if logger.Level >= logrus.TraceLevel {
			// gotask logs
			tasksLogger = &taskLogrusLogger{logger: logger}
			// log callers
			logger.SetReportCaller(true)
		}
	}

	logger.Infof("Version: %v", version.Version())

	// wd
	{
		wd, err := os.Getwd()
		if err != nil {
			logger.WithError(err).Fatal("Failed to get working dir")
		}
		logger.Infof("Working dir: %v", wd)
	}

	// alerter
	var alerter alert.Alerter
	{
		if conf.GCloudAlerts && alert.OnGCE() {
			a, cls, err := alert.NewGCloud("MintSender Sender", logger.WithField("gce_alert", ""))
			if err != nil {
				logger.WithError(err).Fatal("Failed to setup Google Cloud alerts")
			}
			defer cls()
			alerter = a
		} else {
			alerter = alert.NewLogrus(logger.WithField("alert", ""))
		}
	}
	_ = alerter

	// read sender private keys, make senders
	var senderSigners []*signer.Signer
	{
		for i, k := range conf.Wallets {
			b, err := mint.Unpack58(k)
			if err != nil {
				logger.WithError(err).Fatalf("Invalid sender private key at index %v", i)
			}
			sig, err := signer.FromBytes(b)
			if err != nil {
				logger.WithError(err).Fatalf("Invalid sender private key at index %v", i)
			}
			senderSigners = append(senderSigners, sig)
		}
		if len(senderSigners) == 0 {
			logger.Fatal("Add at least one sender private key in Base58")
		}
	}

	// database
	var dao db.DAO
	{
		conf.DB.Prefix = formatPrefix(conf.DB.Prefix, "_")
		if conf.DB.Prefix == "" {
			logger.Fatal("Please specify database table prefix")
		}

		switch strings.ToLower(conf.DB.Driver) {
		case "mysql":
			db, err := mysql.New(conf.DB.DSN, conf.DB.Prefix, false, 16*1024*1024)
			if err != nil {
				logger.WithError(err).Fatal("Failed to setup DB")
			}
			defer db.Close()
			db.DB.DB().SetMaxOpenConns(8)
			db.DB.DB().SetMaxIdleConns(1)
			db.DB.DB().SetConnMaxLifetime(time.Second * 30)
			db.DB.LogMode(logger.Level >= logrus.TraceLevel)
			dao = db
		}

		if !dao.Available() {
			logger.Fatal("Failed to ping DB")
		}
		logger.Info("Connected to DB")

		// migration
		if err := dao.Migrate(); err != nil {
			logger.WithError(err).Fatal("Failed migrate DB")
		}
	}

	// add signers public keys to db to track em after restart
	{
		for _, s := range senderSigners {
			if err := dao.PutWallet(&types.Wallet{
				PublicKey: s.PublicKey(),
			}); err != nil {
				logger.WithError(err).Fatal("Failed to save signer's address to DB")
			}
		}
	}

	// rpc pool
	var rpcPool *rpcpool.Pool
	{
		if len(conf.Nodes) == 0 {
			logger.Fatal("Specify at least one node")
		}
		if p, cls, err := rpcpool.New(conf.Nodes...); err != nil {
			logger.WithError(err).Fatal("Failed to setup RPC pool")
		} else {
			defer cls()
			rpcPool = p
		}
	}

	// get latest block ID
	latestBlockID := new(big.Int)
	{
		ctx, conn, cls, err := rpcPool.Conn()
		if err != nil {
			logger.WithError(err).Fatal("Failed to get latest block ID")
		}

		state, rerr, err := request.GetBlockchainState(ctx, conn)
		if err != nil {
			logger.WithError(err).Fatal("Failed to get latest block ID")
		}
		if rerr != nil {
			logger.WithError(rerr.Err()).Fatal("Failed to get latest block ID")
		}
		latestBlockID.Set(state.BlockCount.Int)
		latestBlockID.Sub(latestBlockID, big.NewInt(1))

		cls()
	}

	// get the earliest unchecked block ID
	var rangerParseFrom *big.Int
	{
		block, ok, err := dao.EarliestBlock()
		if err != nil {
			logger.WithError(err).Fatal("Failed to get earliest block ID")
		}
		if ok {
			rangerParseFrom = block
			if rangerParseFrom.Cmp(new(big.Int)) < 0 {
				logger.Fatalf("Invalid earliest block")
			}
			if rangerParseFrom.Cmp(latestBlockID) > 0 {
				logger.Fatalf("Earliest block ID is greater than current latest block: %v > %v. Blockchain reset?", rangerParseFrom.String(), latestBlockID.String())
			}
			if rangerParseFrom.Cmp(latestBlockID) < 0 {
				logger.Infof("Earliest block to search confirmations from is %v", rangerParseFrom.String())
			}
		}
	}

	// carries parsed transactions
	parsedTX := make(chan *blockparser.Transaction, 256)
	defer close(parsedTX)

	// carries filtered transactions
	filteredTX := make(chan *blockparser.Transaction, 256)
	defer close(filteredTX)

	// carries public keys of wallets to add/remove from transactions filter
	walletToTrack, walletToUntrack := make(chan mint.PublicKey, 32), make(chan mint.PublicKey, 32)
	defer close(walletToTrack)
	defer close(walletToUntrack)

	// carries latest parsed block ID (just a dummy channel here)
	var parsedBlockChan = make(chan *big.Int)
	defer close(parsedBlockChan)

	// fresh block observer
	var blockObserver *blockobserver.Observer
	{
		b, err := blockobserver.New(
			latestBlockID,
			rpcPool,
			parsedTX,
			parsedBlockChan,
			logger.WithField("task", "block_observer"),
		)
		if err != nil {
			logger.WithError(err).Fatal("Failed to setup block observer")
		}
		blockObserver = b
		blockObserverTask, _ = gotask.NewTask("block_observer", blockObserver.Task)
	}

	// blocks range parser
	var blockRanger *blockranger.Ranger
	if rangerParseFrom != nil {
		b, err := blockranger.New(
			rangerParseFrom,
			latestBlockID,
			rpcPool,
			parsedTX,
			parsedBlockChan,
			logger.WithField("task", "block_ranger"),
		)
		if err != nil {
			logger.WithError(err).Fatal("Failed to setup block ranger")
		}
		blockRanger = b
		blockRangerTask, _ = gotask.NewTask("block_ranger", blockRanger.Task)
	}

	// tx filter
	var txFilter *txfilter.Filter
	{
		// type/direction filter
		filter := func(typ transaction.Code, outgoing bool) bool {
			return outgoing && (typ == transaction.TransferAssetTx || typ == transaction.SetWalletTagTx)
		}

		f, err := txfilter.New(
			parsedTX, filteredTX,
			walletToTrack, walletToUntrack,
			filter,
			logger.WithField("task", "tx_filter"),
		)
		if err != nil {
			logger.WithError(err).Fatal("Failed to setup transaction filter")
		}
		txFilter = f
		txFilterTask, _ = gotask.NewTask("tx_filter", txFilter.Task)

		// expose interest on all known signers wallets
		{
			wallets, err := dao.ListWallets()
			if err != nil {
				logger.WithError(err).Fatal("Failed to get all known signers' addresses from DB")
			}
			for _, w := range wallets {
				f.AddWallet(w.PublicKey)
			}
		}
	}

	// tx confirmer
	var txConfirmer *txconfirmer.Confirmer
	{
		c, err := txconfirmer.New(
			filteredTX,
			dao,
			logger.WithField("task", "tx_confirmer"),
		)
		if err != nil {
			logger.WithError(err).Fatal("Failed to setup transaction confirmer")
		}

		txConfirmer = c
		txConfirmerTask, _ = gotask.NewTask("tx_confirmer", txConfirmer.Task)
	}

	// api
	var api *serviceAPI.API
	{
		a, err := serviceAPI.New(
			dao,
			rpcPool,
			logger.WithField("task", "api"),
		)
		if err != nil {
			logger.WithError(err).Fatal("Failed to setup API")
		}
		api = a
	}

	// nats transport
	var natsTransport *serviceNats.Nats
	if conf.API.Nats.URL != "" {
		svc, cls, err := serviceNats.New(
			conf.API.Nats.URL,
			formatPrefix(conf.API.Nats.Prefix, "."),
			api,
			logger.WithField("task", "nats"),
		)
		if err != nil {
			logger.WithError(err).Fatal("Failed to setup Nats transport")
		}
		defer cls()

		natsTransport = svc
		natsTransportTask, _ = gotask.NewTask("nats", svc.Task)
	}

	// http transport
	var httpTransport *serviceHTTP.HTTP
	if conf.API.HTTP.Port > 0 {
		svc, err := serviceHTTP.New(
			conf.API.HTTP.Port,
			api,
			logger.WithField("task", "http"),
		)
		if err != nil {
			logger.WithError(err).Fatal("Failed to setup HTTP transport")
		}

		httpTransport = svc
		httpTransportTask, _ = gotask.NewTask("http", svc.Task)
	}

	// notifier
	{
		var natsIface notifier.NatsTransporter
		if natsTransport != nil {
			natsIface = natsTransport
		}
		var httpIface notifier.HTTPTransporter
		if httpTransport != nil {
			httpIface = httpTransport
		}
		n, err := notifier.New(
			dao,
			natsIface, httpIface,
			logger.WithField("task", "notifier"),
		)
		if err != nil {
			logger.WithError(err).Fatal("Failed to setup notifier")
		}

		notifierTask, _ = gotask.NewTask("notifier", n.Task)
	}

	// tx signer
	var txSigner *txsigner.Signer
	{
		s, err := txsigner.New(
			rpcPool,
			dao,
			senderSigners,
			logger.WithField("task", "tx_signer"),
		)
		if err != nil {
			logger.WithError(err).Fatal("Failed to setup transaction signer")
		}

		txSigner = s
		txSignerTask, _ = gotask.NewTask("tx_signer", s.Task)
	}

	// latest block ID chan consumer
	{
		go func() {
			for {
				_ = <-parsedBlockChan
			}
		}()
	}

	// metrics server
	if conf.Metrics > 0 {
		var ns = "gm"
		var ss = "mintsender_sender"

		// api nats
		if natsTransport != nil {
			m := serviceNats.Metrics{
				RequestDuration: promauto.NewHistogramVec(prometheus.HistogramOpts{
					Name:      "nats_request_duration",
					Help:      "API Nats transport incoming request duration (seconds)",
					Namespace: ns,
					Subsystem: ss,
				}, []string{"method"}),
				NotificationDuration: promauto.NewHistogram(prometheus.HistogramOpts{
					Name:      "nats_notification_duration",
					Help:      "API Nats transport outgoing notification duration (seconds)",
					Namespace: ns,
					Subsystem: ss,
				}),
			}
			natsTransport.AddMetrics(&m)
		}

		// api http
		if httpTransport != nil {
			m := serviceHTTP.Metrics{
				RequestDuration: promauto.NewHistogramVec(prometheus.HistogramOpts{
					Name:      "http_request_duration",
					Help:      "API HTTP transport incoming request duration (seconds)",
					Namespace: ns,
					Subsystem: ss,
				}, []string{"method"}),
				NotificationDuration: promauto.NewHistogram(prometheus.HistogramOpts{
					Name:      "http_notification_duration",
					Help:      "API HTTP transport outgoing notification duration (seconds)",
					Namespace: ns,
					Subsystem: ss,
				}),
			}
			httpTransport.AddMetrics(&m)
		}

		// block parser
		{
			m := blockparser.Metrics{
				RequestDuration: promauto.NewHistogram(prometheus.HistogramOpts{
					Name:      "blockparser_request_duration",
					Help:      "Block parser delivery duration (seconds)",
					Namespace: ns,
					Subsystem: ss,
				}),
				ParsingDuration: promauto.NewHistogram(prometheus.HistogramOpts{
					Name:      "blockparser_parsing_duration",
					Help:      "Block parser reading duration (seconds)",
					Namespace: ns,
					Subsystem: ss,
				}),
			}
			blockObserver.AddMetrics(&m)
			if blockRanger != nil {
				blockRanger.AddMetrics(&m)
			}
		}

		// tx filter
		{
			m := txfilter.Metrics{
				ROIWallets: promauto.NewGauge(prometheus.GaugeOpts{
					Name:      "txfilter_roi_wallets",
					Help:      "Transaction filter wallets count in ROI",
					Namespace: ns,
					Subsystem: ss,
				}),
				TxVolume: promauto.NewCounterVec(prometheus.CounterOpts{
					Name:      "txfilter_txvolume",
					Help:      "Amount of token sent",
					Namespace: ns,
					Subsystem: ss,
				}, []string{"token"}),
			}
			txFilter.AddMetrics(&m)
		}

		// tx signer
		{
			m := txsigner.Metrics{
				Balance: promauto.NewGaugeVec(prometheus.GaugeOpts{
					Name:      "txsigner_balance",
					Help:      "Sender wallets balance",
					Namespace: ns,
					Subsystem: ss,
				}, []string{"wallet", "token"}),
				Queue: promauto.NewGauge(prometheus.GaugeOpts{
					Name:      "txsigner_queue",
					Help:      "Transaction signer queue size",
					Namespace: ns,
					Subsystem: ss,
				}),
			}
			txSigner.AddMetrics(&m)
		}

		m := metrics.New(uint16(conf.Metrics), logger.WithField("task", "metrics"))
		metricsTask, _ = gotask.NewTask("metrics", m.Task)
	}

	// handle termination signal
	go func() {
		sigchan := make(chan os.Signal, 1)
		signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)
		<-sigchan
		logger.Info("Stop signal received")
		onStop()
	}()

	group = gotask.NewGroup("main")
	group.Log(tasksLogger)
	tasks := []*gotask.Task{
		metricsTask,
		blockObserverTask,
		blockRangerTask,
		txFilterTask,
		txConfirmerTask,
		natsTransportTask,
		httpTransportTask,
		notifierTask,
		txSignerTask,
	}
	for _, t := range tasks {
		if t != nil {
			t.Log(tasksLogger)
			if err := group.Add(t); err != nil {
				logger.WithError(err).Fatal("Failed to add task", t.Tag())
			}
		}
	}

	logger.Debug("Launch tasks")
	if err := group.Run(); err != nil {
		logger.WithError(err).Fatal("Failed to run tasks")
	}

	logger.Debug("Awaiting stop signal")
	group.Wait()

	logger.Info("Graceful stop")
}

func onStop() {
	stopWait(txSignerTask)
	stopWait(notifierTask)
	stopWait(natsTransportTask)
	stopWait(httpTransportTask)
	stopWait(blockObserverTask)
	stopWait(blockRangerTask)
	stopWait(txFilterTask)
	stopWait(txConfirmerTask)
	stopWait(metricsTask)
	group.Stop()
}

// ---

type config struct {
	Log struct {
		Level string `yaml:"level"`
		Color bool   `yaml:"color"`
		JSON  bool   `yaml:"json"`
	} `yaml:"log"`

	API struct {
		HTTP struct {
			Port uint `yaml:"port"`
		} `yaml:"http"`

		Nats struct {
			URL    string `yaml:"url"`
			Prefix string `yaml:"prefix"`
		} `yaml:"nats"`
	} `yaml:"api"`

	DB struct {
		Driver string `yaml:"driver"`
		DSN    string `yaml:"dsn"`
		Prefix string `yaml:"prefix"`
	} `yaml:"db"`

	Metrics      uint     `yaml:"metrics"`
	GCloudAlerts bool     `yaml:"gcloud_alerts"`
	Nodes        []string `yaml:"nodes"`
	Wallets      []string `yaml:"wallets"`
}

// ---

func stopWait(t *gotask.Task) {
	if t == nil {
		return
	}
	if token, err := group.TokenOf(t); err != nil {
		logger.WithError(err).WithField("task", t.Tag()).Error("Failed to stop task")
	} else {
		token.Stop()
		if waiter, err := group.WaiterOf(t); err != nil {
			logger.WithError(err).WithField("task", t.Tag()).Error("Failed to wait task")
		} else {
			waiter.Wait()
		}
	}
}

// ---

type taskLogrusLogger struct {
	gotask.Logger
	logger *logrus.Logger
}

func (l *taskLogrusLogger) Log(args ...interface{}) {
	l.logger.Traceln(args...)
}

// ---

type logrusErrorHook struct {
	addError func(source string, level logrus.Level)
}

func (h *logrusErrorHook) Levels() []logrus.Level {
	return []logrus.Level{logrus.FatalLevel, logrus.ErrorLevel, logrus.WarnLevel}
}

func (h *logrusErrorHook) Fire(e *logrus.Entry) error {
	source, ok := e.Data["task"]
	if !ok {
		source = "main"
	}
	go h.addError(fmt.Sprint(source), e.Level)
	return nil
}

// ---

func formatPrefix(s, delim string) string {
	charz := []rune(strings.TrimSpace(s))
	length := len(charz)
	switch {
	case length == 0:
		return ""
	case unicode.IsDigit(charz[length-1]) || !unicode.IsNumber(charz[length-1]):
		return string(charz) + delim
	default:
		return string(charz)
	}
}
