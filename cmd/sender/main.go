package main

import (
	"encoding/json"
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
	"github.com/void616/gm-mint-sender/internal/blockobserver"
	"github.com/void616/gm-mint-sender/internal/blockparser"
	"github.com/void616/gm-mint-sender/internal/blockranger"
	"github.com/void616/gm-mint-sender/internal/metrics"
	"github.com/void616/gm-mint-sender/internal/rpcpool"
	"github.com/void616/gm-mint-sender/internal/sender/db"
	"github.com/void616/gm-mint-sender/internal/sender/db/mysql"
	"github.com/void616/gm-mint-sender/internal/sender/db/types"
	"github.com/void616/gm-mint-sender/internal/sender/notifier"
	"github.com/void616/gm-mint-sender/internal/sender/senderservice"
	serviceNats "github.com/void616/gm-mint-sender/internal/sender/transport/nats"
	"github.com/void616/gm-mint-sender/internal/sender/txconfirmer"
	"github.com/void616/gm-mint-sender/internal/sender/txsigner"
	"github.com/void616/gm-mint-sender/internal/txfilter"
	"github.com/void616/gm-mint-sender/internal/version"
	sumuslib "github.com/void616/gm-sumuslib"
	"github.com/void616/gm-sumuslib/signer"
	"github.com/void616/gm-sumusrpc/rpc"
	"github.com/void616/gotask"
	gormigrate "gopkg.in/gormigrate.v1"
)

var (
	logger            *logrus.Logger
	group             *gotask.Group
	blockObserverTask *gotask.Task
	blockRangerTask   *gotask.Task
	txFilterTask      *gotask.Task
	txConfirmerTask   *gotask.Task
	natsTransportTask *gotask.Task
	notifierTask      *gotask.Task
	txSignerTask      *gotask.Task
	metricsTask       *gotask.Task
)

func main() {

	// flags
	var (
		// log
		logLevel   = flag.String("log", "info", "Log level: fatal|error|warn|info|debug|trace")
		logJSON    = flag.Bool("json", false, "Log in Json format")
		logNoColor = flag.Bool("nocolor", false, "Disable colorful log")
		// keys
		senderKeysFile = flag.String("keys", "keys.json", "Path to file contains set of sender private keys")
		// nodes
		sumusNodes flagArray
		// db
		dbDSN         = flag.String("dsn", "", "Database connection string")
		dbTablePrefix = flag.String("table", "sender", "Database table prefix")
		// nats
		natsURL        = flag.String("nats", "localhost:4222", "Nats server endpoint")
		natsSubjPrefix = flag.String("nats-subj", "", "Prefix for Nats messages subject")
		// metrics
		metricsPort = flag.Uint("metrics", 0, "Prometheus port, i.e. 2112")
	)
	flag.Var(&sumusNodes, "node", "Sumus RPC encpoints, i.e. 127.0.0.1:4010")
	flag.Parse()

	// logging
	logger = logrus.New()
	var tasksLogger gotask.Logger
	{
		lvl, err := logrus.ParseLevel(*logLevel)
		if err != nil {
			logger.WithError(err).Fatal("Failed to parse logger level")
		}
		logger.Out = os.Stdout
		logger.Level = lvl

		// format
		switch {
		case *logJSON:
			logger.Formatter = &logrus.JSONFormatter{
				TimestampFormat: time.RFC3339,
			}
		default:
			logger.Formatter = &logrus.TextFormatter{
				FullTimestamp:   true,
				ForceColors:     !*logNoColor,
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

	// read sender private keys, make senders
	var senderSigners []*signer.Signer
	{
		b, err := ioutil.ReadFile(*senderKeysFile)
		if err != nil {
			logger.WithError(err).Fatal("Failed to read sender keys file")
		}
		cfg := &senderKeysConfig{}
		if err := json.Unmarshal(b, &cfg); err != nil {
			logger.WithError(err).Fatal("Failed to unmarshal sender keys")
		}
		for i, k := range cfg.Keys {
			b, err := sumuslib.Unpack58(k)
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
		*dbTablePrefix = formatPrefix(*dbTablePrefix, "_")
		if *dbTablePrefix == "" {
			logger.Fatal("Please specify database table prefix")
		}

		db, err := mysql.New(*dbDSN, *dbTablePrefix, false, 16*1024*1024)
		if err != nil {
			logger.WithError(err).Fatal("Failed to setup DB")
		}
		defer db.Close()
		db.DB.DB().SetMaxOpenConns(8)
		db.DB.DB().SetMaxIdleConns(1)
		db.DB.DB().SetConnMaxLifetime(time.Second * 30)

		dao = db
		if !dao.Available() {
			logger.WithError(err).Fatal("Failed to ping DB")
		}
		logger.Info("Connected to DB")

		db.DB.LogMode(logger.Level >= logrus.TraceLevel)

		// migration
		{
			opts := gormigrate.DefaultOptions
			opts.TableName = *dbTablePrefix + "dbmigrations"
			mig := gormigrate.New(db.DB, opts, mysql.Migrations)
			if err := mig.Migrate(); err != nil {
				logger.WithError(err).Fatal("Failed to apply DB migration")
			}
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

	// metrics
	var mtxSenderServiceRequestDuration *prometheus.SummaryVec
	var mtxNatsRequestDuration *prometheus.SummaryVec
	var mtxTaskDuration *prometheus.SummaryVec
	var mtxQueueGauge *prometheus.GaugeVec
	var mtxTxVolumeCounter *prometheus.CounterVec
	var mtxWalletBalanceGauge *prometheus.GaugeVec
	var mtxErrorCounter *prometheus.CounterVec
	if *metricsPort != 0 {
		ns := "gmmintsender"
		ss := "sender"

		mtxSenderServiceRequestDuration = promauto.NewSummaryVec(prometheus.SummaryOpts{
			Name:      "sendersvc_request_duration",
			Help:      "Sender service request duration (seconds)",
			Namespace: ns, Subsystem: ss,
		}, []string{"method"})

		mtxNatsRequestDuration = promauto.NewSummaryVec(prometheus.SummaryOpts{
			Name:      "natsapi_request_duration",
			Help:      "Nats API request duration",
			Namespace: ns, Subsystem: ss,
		}, []string{"method"})

		mtxTaskDuration = promauto.NewSummaryVec(prometheus.SummaryOpts{
			Name:      "task_duration",
			Help:      "Internal task duration (seconds)",
			Namespace: ns, Subsystem: ss,
		}, []string{"task"})

		mtxQueueGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name:      "queue_size",
			Help:      "Internal queue size",
			Namespace: ns, Subsystem: ss,
		}, []string{"queue"})

		mtxTxVolumeCounter = promauto.NewCounterVec(prometheus.CounterOpts{
			Name:      "tx_volume",
			Help:      "Volume of sent transactions",
			Namespace: ns, Subsystem: ss,
		}, []string{"token"})

		mtxWalletBalanceGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name:      "wallets",
			Help:      "Sender wallet balance",
			Namespace: ns, Subsystem: ss,
		}, []string{"sender", "token"})

		mtxErrorCounter = promauto.NewCounterVec(prometheus.CounterOpts{
			Name:      "errors",
			Help:      "Error counter",
			Namespace: ns, Subsystem: ss,
		}, []string{"source", "level"})

		hook := &logrusErrorHook{
			addError: func(source string, level logrus.Level) {
				mtxErrorCounter.WithLabelValues(source, level.String()).Add(1)
			},
		}
		logger.AddHook(hook)
	}

	// rpc pool
	var rpcPool *rpcpool.Pool
	{
		if len(sumusNodes) == 0 {
			logger.Fatal("Specify at least one Sumus node with --node flag")
		}
		if p, cls, err := rpcpool.New(sumusNodes...); err != nil {
			logger.WithError(err).Fatal("Failed to setup Sumus RPC pool")
		} else {
			defer cls()
			rpcPool = p
		}
	}

	// get latest block ID
	latestBlockID := new(big.Int)
	{
		c, err := rpcPool.Get()
		if err != nil {
			logger.WithError(err).Fatal("Failed to get latest block ID")
		}
		state, code, err := rpc.BlockchainState(c.Conn())
		if err != nil {
			logger.WithError(err).Fatal("Failed to get latest block ID")
		}
		if code != rpc.ECSuccess {
			logger.WithError(fmt.Errorf("node code %v", code)).Fatal("Failed to get latest block ID")
		}
		latestBlockID.Set(state.BlockCount)
		latestBlockID.Sub(latestBlockID, big.NewInt(1))
		c.Close()
	}

	// get the earliest unchecked block ID
	var rangerParseFrom *big.Int
	{
		block, empty, err := dao.EarliestBlock()
		if err != nil {
			logger.WithError(err).Fatal("Failed to get earliest block ID")
		}
		if !empty {
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
	walletToTrack, walletToUntrack := make(chan sumuslib.PublicKey, 32), make(chan sumuslib.PublicKey, 32)
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
			mtxTaskDuration, mtxQueueGauge,
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
			mtxTaskDuration,
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
		filter := func(typ sumuslib.Transaction, outgoing bool) bool {
			return typ == sumuslib.TransactionTransferAssets && outgoing
		}

		f, err := txfilter.New(
			parsedTX, filteredTX,
			walletToTrack, walletToUntrack,
			filter,
			nil, mtxTxVolumeCounter, mtxTaskDuration, mtxQueueGauge,
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
			mtxTaskDuration,
			logger.WithField("task", "tx_confirmer"),
		)
		if err != nil {
			logger.WithError(err).Fatal("Failed to setup transaction confirmer")
		}

		txConfirmer = c
		txConfirmerTask, _ = gotask.NewTask("tx_confirmer", txConfirmer.Task)
	}

	// wallet service
	var senderService *senderservice.Service
	{
		s, err := senderservice.New(
			dao,
			mtxSenderServiceRequestDuration,
			logger.WithField("task", "sender_service"),
		)
		if err != nil {
			logger.WithError(err).Fatal("Failed to setup sender service")
		}
		senderService = s
	}

	// nats transport
	var natsTransport *serviceNats.Service
	{
		n, cls, err := serviceNats.New(
			*natsURL,
			formatPrefix(*natsSubjPrefix, "."),
			senderService,
			mtxNatsRequestDuration, mtxTaskDuration, mtxQueueGauge,
			logger.WithField("task", "nats_transport"),
		)
		if err != nil {
			logger.WithError(err).Fatal("Failed to setup Nats transport")
		}
		defer cls()

		natsTransport = n
		natsTransportTask, _ = gotask.NewTask("nats_transport", n.Task)
	}

	// notifier
	{
		n, err := notifier.New(
			dao,
			natsTransport,
			mtxTaskDuration, mtxQueueGauge,
			logger.WithField("task", "notifier"),
		)
		if err != nil {
			logger.WithError(err).Fatal("Failed to setup notifier")
		}

		notifierTask, _ = gotask.NewTask("notifier", n.Task)
	}

	// tx signer
	{
		s, err := txsigner.New(
			rpcPool,
			dao,
			senderSigners,
			mtxWalletBalanceGauge, mtxTaskDuration, mtxQueueGauge,
			logger.WithField("task", "tx_signer"),
		)
		if err != nil {
			logger.WithError(err).Fatal("Failed to setup transaction signer")
		}

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
	if *metricsPort > 0 {
		m := metrics.New(uint16(*metricsPort), logger.WithField("task", "metrics"))
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
	stopWait(blockObserverTask)
	stopWait(blockRangerTask)
	stopWait(txFilterTask)
	stopWait(txConfirmerTask)
	stopWait(metricsTask)
	group.Stop()
}

// ---

type flagArray []string

func (i *flagArray) String() string {
	return strings.Join(*i, " ")
}

func (i *flagArray) Set(value string) error {
	*i = append(*i, value)
	return nil
}

// ---

type senderKeysConfig struct {
	Keys []string `json:"keys,omitempty"`
}

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
