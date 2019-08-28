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
	"github.com/void616/gm-mint-sender/internal/blockobserver"
	"github.com/void616/gm-mint-sender/internal/blockparser"
	"github.com/void616/gm-mint-sender/internal/blockranger"
	"github.com/void616/gm-mint-sender/internal/metrics"
	"github.com/void616/gm-mint-sender/internal/rpcpool"
	"github.com/void616/gm-mint-sender/internal/txfilter"
	"github.com/void616/gm-mint-sender/internal/version"
	"github.com/void616/gm-mint-sender/internal/watcher/db"
	"github.com/void616/gm-mint-sender/internal/watcher/db/mysql"
	"github.com/void616/gm-mint-sender/internal/watcher/notifier"
	serviceNats "github.com/void616/gm-mint-sender/internal/watcher/transport/nats"
	"github.com/void616/gm-mint-sender/internal/watcher/txsaver"
	"github.com/void616/gm-mint-sender/internal/watcher/walletservice"
	sumuslib "github.com/void616/gm-sumuslib"
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
	txSaverTask       *gotask.Task
	natsTransportTask *gotask.Task
	notifierTask      *gotask.Task
	metricsTask       *gotask.Task
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

	// metrics
	var mtxWalletServiceRequestDuration *prometheus.SummaryVec
	var mtxNatsRequestDuration *prometheus.SummaryVec
	var mtxTaskDuration *prometheus.SummaryVec
	var mtxQueueGauge *prometheus.GaugeVec
	var mtxTxVolumeCounter *prometheus.CounterVec
	var mtxROIWalletsGauge prometheus.Gauge
	var mtxErrorCounter *prometheus.CounterVec
	if *metricsPort != 0 {
		ns := "gmmintsender"
		ss := "watcher"

		mtxWalletServiceRequestDuration = promauto.NewSummaryVec(prometheus.SummaryOpts{
			Name:      "walletsvc_request_duration",
			Help:      "Wallet service requrest duration (seconds)",
			Namespace: ns, Subsystem: ss,
		}, []string{"method"})

		mtxNatsRequestDuration = promauto.NewSummaryVec(prometheus.SummaryOpts{
			Name:      "natsapi_request_duration",
			Help:      "Nats API requrest duration (seconds)",
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
			Help:      "Volume of received transactions",
			Namespace: ns, Subsystem: ss,
		}, []string{"token"})

		mtxROIWalletsGauge = promauto.NewGauge(prometheus.GaugeOpts{
			Name:      "wallets",
			Help:      "ROI wallets counter",
			Namespace: ns, Subsystem: ss,
		})

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

	// parser transactions chan
	parsedTX := make(chan *blockparser.Transaction, 256)
	defer close(parsedTX)

	// filtered transactions chan
	filteredTX := make(chan *blockparser.Transaction, 256)
	defer close(filteredTX)

	// a pair of chans to add/remove wallets to/from filter
	walletToAdd, walletToRemove := make(chan sumuslib.PublicKey, 32), make(chan sumuslib.PublicKey, 32)
	defer close(walletToAdd)
	defer close(walletToRemove)

	// fresh block observer
	var blockObserver *blockobserver.Observer
	{
		b, err := blockobserver.New(
			latestBlockID,
			rpcPool,
			parsedTX,
			mtxTaskDuration, mtxQueueGauge,
			logger.WithField("task", "block_observer"),
		)
		if err != nil {
			logger.WithError(err).Fatal("Failed to setup block observer")
		}
		blockObserver = b
		blockObserverTask, _ = gotask.NewTask("block_observer", blockObserver.Task)
	}

	// range of blocks parser
	var blockRanger *blockranger.Ranger
	if *rangerParseFrom != "" {
		from, ok := big.NewInt(0).SetString(*rangerParseFrom, 10)
		if !ok || from.Cmp(new(big.Int)) < 0 {
			logger.Fatal("Invalid range start")
		}
		if from.Cmp(latestBlockID) > 0 {
			logger.Fatalf("Invalid range start. Current latest block ID is %v", latestBlockID.String())
		}

		b, err := blockranger.New(
			from,
			latestBlockID,
			rpcPool,
			parsedTX,
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
			return typ == sumuslib.TransactionTransferAssets && !outgoing
		}

		f, err := txfilter.New(
			parsedTX,
			filteredTX,
			walletToAdd,
			walletToRemove,
			filter,
			mtxROIWalletsGauge, mtxTxVolumeCounter, mtxTaskDuration, mtxQueueGauge,
			logger.WithField("task", "tx_filter"),
		)
		if err != nil {
			logger.WithError(err).Fatal("Failed to setup transaction filter")
		}
		txFilter = f
		txFilterTask, _ = gotask.NewTask("tx_filter", txFilter.Task)

		// add wallets to the filter
		if wallets, err := dao.ListWallets(); err != nil {
			logger.WithError(err).Fatal("Failed to get wallets list from DB")
		} else {
			for _, w := range wallets {
				f.AddWallet(w.PublicKey)
			}
		}
	}

	// tx saver
	var txSaver *txsaver.Saver
	{
		s, err := txsaver.New(
			filteredTX,
			dao,
			mtxTaskDuration,
			logger.WithField("task", "tx_saver"),
		)
		if err != nil {
			logger.WithError(err).Fatal("Failed to setup transaction saver")
		}

		txSaver = s
		txSaverTask, _ = gotask.NewTask("tx_saver", txSaver.Task)
	}

	// wallet service
	var walletService *walletservice.Service
	{
		s, err := walletservice.New(
			walletToAdd,
			walletToRemove,
			dao,
			mtxWalletServiceRequestDuration,
			logger.WithField("task", "wallet_service"),
		)
		if err != nil {
			logger.WithError(err).Fatal("Failed to setup wallet service")
		}
		walletService = s
	}

	// nats transport
	var natsTransport *serviceNats.Service
	{
		n, cls, err := serviceNats.New(
			*natsURL,
			formatPrefix(*natsSubjPrefix, "."),
			walletService,
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

	// refilling notifier
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
