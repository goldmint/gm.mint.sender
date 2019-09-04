package txsigner

import (
	"fmt"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"github.com/void616/gm-mint-sender/internal/mint/rpcpool"
	"github.com/void616/gm-mint-sender/internal/sender/db"
	sumuslib "github.com/void616/gm-sumuslib"
	"github.com/void616/gm-sumuslib/amount"
	sumusSigner "github.com/void616/gm-sumuslib/signer"
	"github.com/void616/gm-sumusrpc/rpc"
)

const itemsPerShot = 50
const staleAfterBlocks = 1

// Signer signs and sends transactions
type Signer struct {
	logger  *logrus.Entry
	pool    *rpcpool.Pool
	signers map[sumuslib.PublicKey]*SignerData
	dao     db.DAO
	metrics *Metrics
}

// SignerData describes particular signer
type SignerData struct {
	signer      *sumusSigner.Signer
	public      sumuslib.PublicKey
	nonce       uint64
	gold        *amount.Amount
	mnt         *amount.Amount
	emitter     bool
	signedCount uint64
	failed      bool
}

// New Signer instance
func New(
	pool *rpcpool.Pool,
	dao db.DAO,
	signers []*sumusSigner.Signer,
	logger *logrus.Entry,
) (*Signer, error) {

	// get rpc connection
	conn, err := pool.Get()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// make a map of signers with some extra data required in runtime
	signerz := make(map[sumuslib.PublicKey]*SignerData)
	for _, ss := range signers {
		pubkey := ss.PublicKey()

		// get wallet state from the network
		walletState, code, err := rpc.WalletState(conn.Conn(), pubkey.String())
		if err != nil {
			return nil, err
		}
		if code != rpc.ECSuccess {
			return nil, fmt.Errorf("node code %v", code)
		}
		emitter := false
		for _, t := range walletState.Tags {
			if t == sumuslib.WalletTagEmission.String() {
				emitter = true
				break
			}
		}

		// check db for a greater nonce
		dbnonce, err := dao.LatestSenderNonce(pubkey)
		if err != nil {
			return nil, err
		}

		// use the greatest
		nonce := walletState.ApprovedNonce
		if dbnonce > nonce {
			nonce = dbnonce
		}

		signerz[pubkey] = &SignerData{
			signer:      ss,
			public:      pubkey,
			nonce:       nonce,
			gold:        amount.FromAmount(walletState.Balance.Gold),
			mnt:         amount.FromAmount(walletState.Balance.Mnt),
			emitter:     emitter,
			signedCount: 0,
			failed:      false,
		}

		logGold, logMnt := strconv.FormatFloat(walletState.Balance.Gold.Float64(), 'f', 6, 64), strconv.FormatFloat(walletState.Balance.Mnt.Float64(), 'f', 6, 64)
		if emitter {
			logGold, logMnt = "emitter", "emitter"
		}
		logger.
			WithField("net_nonce", walletState.ApprovedNonce).
			WithField("db_nonce", dbnonce).
			WithField("gold", logGold).
			WithField("mnt", logMnt).
			Infof("Signer %v prepared", pubkey.StringMask())
	}

	s := &Signer{
		logger:  logger,
		dao:     dao,
		pool:    pool,
		signers: signerz,
	}
	return s, nil
}

// Metrics data
type Metrics struct {
	Balance *prometheus.GaugeVec
	Queue   prometheus.Gauge
}

// AddMetrics adds metrics counters and should be called before service launch
func (s *Signer) AddMetrics(m *Metrics) {
	s.metrics = m

	if m != nil {
		// metrics
		for pub, sig := range s.signers {
			m.Balance.WithLabelValues(pub.String(), "gold").Set(sig.gold.Float64())
			m.Balance.WithLabelValues(pub.String(), "mnt").Set(sig.mnt.Float64())
		}
	}
}
