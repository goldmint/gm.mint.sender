package main

import (
	"flag"
	"fmt"
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
	"github.com/void616/gm-mint-sender/internal/metrics"
	"github.com/void616/gm-mint-sender/internal/mint/blockobserver"
	"github.com/void616/gm-mint-sender/internal/mint/blockparser"
	"github.com/void616/gm-mint-sender/internal/mint/blockranger"
	"github.com/void616/gm-mint-sender/internal/mint/rpcpool"
	"github.com/void616/gm-mint-sender/internal/mint/txfilter"
	"github.com/void616/gm-mint-sender/internal/version"
	serviceAPI "github.com/void616/gm-mint-sender/internal/watcher/api"
	apiModels "github.com/void616/gm-mint-sender/internal/watcher/api/model"
	serviceNats "github.com/void616/gm-mint-sender/internal/watcher/api/nats"
	"github.com/void616/gm-mint-sender/internal/watcher/db"
	"github.com/void616/gm-mint-sender/internal/watcher/db/mysql"
	"github.com/void616/gm-mint-sender/internal/watcher/db/types"
	"github.com/void616/gm-mint-sender/internal/watcher/notifier"
	"github.com/void616/gm-mint-sender/internal/watcher/txsaver"
	sumuslib "github.com/void616/gm-sumuslib"
	"github.com/void616/gm-sumusrpc/rpc"
	"github.com/void616/gotask"
	gormigrate "gopkg.in/gormigrate.v1"
)

var (
	logger              *logrus.Logger
	group               *gotask.Group
	blockObserverTask   *gotask.Task
	blockRangerTask     *gotask.Task
	txFilterTask        *gotask.Task
	txSaverTask         *gotask.Task
	natsTransportTask   *gotask.Task
	notifierTask        *gotask.Task
	metricsTask         *gotask.Task
	lastParsedBlockTask *gotask.Task
)

func main() {

	// flags
	var (
		// log
		logLevel   = flag.String("log", "info", "Log level: fatal|error|warn|info|debug|trace")
		logJSON    = flag.Bool("json", false, "Log in Json format")
		logNoColor = flag.Bool("nocolor", false, "Disable colorful log")
		// nodes
		sumusNodes      flagArray
		rangerParseFrom = flag.String("parse-from", "", "Run blocks range parser from the specific block (inclusive) to the current latest block")
		// db
		dbDSN         = flag.String("dsn", "", "Database connection string")
		dbTablePrefix = flag.String("table", "watcher", "Database table prefix")
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

	// carries latest parsed block ID
	var parsedBlockChan = make(chan *big.Int)
	defer close(parsedBlockChan)

	// get latest block ID (network)
	var latestBlockID = new(big.Int)
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

	// get latest parsed block ID (DB)
	var latestParsedBlockID = new(big.Int)
	{
		latestParsed, err := dao.GetSetting(types.SettingLatestBlock, "0")
		if err != nil {
			logger.WithError(err).Fatal("Failed to read latest parsed block ID")
		}
		if _, ok := latestParsedBlockID.SetString(latestParsed, 10); !ok || latestParsedBlockID.Cmp(new(big.Int)) < 0 {
			logger.WithError(err).Fatal("Failed to set latest parsed block ID")
		}
		if latestParsedBlockID.Cmp(new(big.Int)) == 0 {
			latestParsedBlockID.Set(latestBlockID)
			dao.PutSetting(types.SettingLatestBlock, latestParsedBlockID.String())
		}
	}

	// block ID, explicitly set via args, to parse from
	var userParseFromBlockID *big.Int
	if *rangerParseFrom != "" {
		userParseFromBlockID = new(big.Int)
		if _, ok := userParseFromBlockID.SetString(*rangerParseFrom, 10); !ok || userParseFromBlockID.Cmp(new(big.Int)) < 0 {
			logger.Fatal("Failed to set blocks range from app argument")
		}
	}

	// carries parsed transactions
	var parsedTX = make(chan *blockparser.Transaction, 256)
	defer close(parsedTX)

	// carries filtered transactions
	var filteredTX = make(chan *blockparser.Transaction, 256)
	defer close(filteredTX)

	// carries public keys of wallets to add/remove from transactions filter
	var walletToTrack, walletToUntrack = make(chan sumuslib.PublicKey, 256), make(chan sumuslib.PublicKey, 256)
	defer close(walletToTrack)
	defer close(walletToUntrack)

	// carries wallet:service pairs to add/remove from transactions saver
	var walletSubs = make(chan apiModels.WalletSub, 512)
	defer close(walletSubs)

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

	// blocks range parser (from DB or from arg)
	var blockRanger *blockranger.Ranger
	{
		// from: set from DB
		from := new(big.Int).Add(latestParsedBlockID, big.NewInt(1))
		// from: user has specified lesser ID
		if userParseFromBlockID != nil {
			if from.Cmp(userParseFromBlockID) > 0 {
				from.Set(userParseFromBlockID)
			}
		}
		// latest >= from
		if latestBlockID.Cmp(from) >= 0 {
			b, err := blockranger.New(
				from,
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
	}

	// tx filter
	var txFilter *txfilter.Filter
	{
		// type/direction filter
		filter := func(typ sumuslib.Transaction, outgoing bool) bool {
			return typ == sumuslib.TransactionTransferAssets && !outgoing
		}

		f, err := txfilter.New(
			parsedTX,
			filteredTX,
			walletToTrack,
			walletToUntrack,
			filter,
			logger.WithField("task", "tx_filter"),
		)
		if err != nil {
			logger.WithError(err).Fatal("Failed to setup transaction filter")
		}
		txFilter = f
		txFilterTask, _ = gotask.NewTask("tx_filter", txFilter.Task)
	}

	// tx saver
	var txSaver *txsaver.Saver
	{
		s, err := txsaver.New(
			filteredTX,
			walletSubs,
			walletToUntrack,
			dao,
			logger.WithField("task", "tx_saver"),
		)
		if err != nil {
			logger.WithError(err).Fatal("Failed to setup transaction saver")
		}

		txSaver = s
		txSaverTask, _ = gotask.NewTask("tx_saver", txSaver.Task)
	}

	// wallet service
	var api *serviceAPI.API
	{
		a, err := serviceAPI.New(
			walletToTrack,
			walletSubs,
			dao,
			logger.WithField("task", "api"),
		)
		if err != nil {
			logger.WithError(err).Fatal("Failed to setup API")
		}
		api = a
	}

	// load all the tracking wallets and subscribed services
	{
		wallets, err := dao.ListWallets()
		if err != nil {
			logger.WithError(err).Fatal("Failed to get wallets list from DB")
		}
		for _, w := range wallets {
			txFilter.AddWallet(w.PublicKey)
			txSaver.AddWalletSubs(w.PublicKey, w.Services...)
		}
	}

	// nats transport
	var natsTransport *serviceNats.Nats
	{
		n, cls, err := serviceNats.New(
			*natsURL,
			formatPrefix(*natsSubjPrefix, "."),
			api,
			logger.WithField("task", "nats"),
		)
		if err != nil {
			logger.WithError(err).Fatal("Failed to setup Nats transport")
		}
		defer cls()

		natsTransport = n
		natsTransportTask, _ = gotask.NewTask("nats", n.Task)
	}

	// refilling notifier
	{
		n, err := notifier.New(
			dao,
			natsTransport,
			logger.WithField("task", "notifier"),
		)
		if err != nil {
			logger.WithError(err).Fatal("Failed to setup notifier")
		}

		notifierTask, _ = gotask.NewTask("notifier", n.Task)
	}

	// metrics server
	if *metricsPort > 0 {
		var ns = "gm"
		var ss = "mintsender_watcher"

		// api nats
		{
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
			blockRanger.AddMetrics(&m)
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

		m := metrics.New(uint16(*metricsPort), logger.WithField("task", "metrics"))
		metricsTask, _ = gotask.NewTask("metrics", m.Task)
	}

	// last parsed block ID saver
	{
		lastParsedBlockTask, _ = gotask.NewTask("blockid_saver", func(token *gotask.Token, arg ...interface{}) {
			var oldid = arg[0].(*big.Int)
			var newid = new(big.Int)
			var save = func() {
				if newid.Cmp(oldid) > 0 {
					dao.PutSetting(types.SettingLatestBlock, newid.String())
					oldid.Set(newid)
				}
			}
			for !token.Stopped() {
				select {
				case id := <-parsedBlockChan:
					if id.Cmp(newid) > 0 {
						newid.Set(id)
						if newid.Cmp(new(big.Int).Add(oldid, big.NewInt(10))) > 0 {
							save()
						}
					}
				case <-time.After(time.Millisecond * 250):
					save()
				}
			}
			save()
		}, new(big.Int).Set(latestParsedBlockID))
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
		lastParsedBlockTask,
		metricsTask,
		blockObserverTask,
		blockRangerTask,
		txFilterTask,
		txSaverTask,
		natsTransportTask,
		notifierTask,
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
	stopWait(notifierTask)
	stopWait(natsTransportTask)
	stopWait(blockObserverTask)
	stopWait(blockRangerTask)
	stopWait(txFilterTask)
	stopWait(txSaverTask)
	stopWait(metricsTask)
	stopWait(lastParsedBlockTask)
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
