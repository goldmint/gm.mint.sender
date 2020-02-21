package txsigner

import (
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	mint "github.com/void616/gm.mint"
	"github.com/void616/gm.mint.rpc/request"
	"github.com/void616/gm.mint.sender/internal/mint/rpcpool"
	"github.com/void616/gm.mint.sender/internal/sender/db"
	"github.com/void616/gm.mint/amount"
	"github.com/void616/gm.mint/signer"
)

const itemsPerShot = 25
const staleAfterBlocks = 1

// Signer signs and sends transactions
type Signer struct {
	logger  *logrus.Entry
	pool    *rpcpool.Pool
	signers map[mint.PublicKey]*SignerData
	dao     db.DAO
	metrics *Metrics
}

// SignerData describes particular signer
type SignerData struct {
	signer      *signer.Signer
	public      mint.PublicKey
	nonce       uint64
	gold        *amount.Amount
	mnt         *amount.Amount
	emitter     bool
	approver    bool
	signedCount uint64
}

// New Signer instance
func New(
	pool *rpcpool.Pool,
	dao db.DAO,
	signers []*signer.Signer,
	logger *logrus.Entry,
) (*Signer, error) {

	// get rpc connection
	ctx, conn, cls, err := pool.Conn()
	if err != nil {
		return nil, err
	}
	defer cls()

	// make a map of signers with some extra data required in runtime
	signerz := make(map[mint.PublicKey]*SignerData)
	for _, ss := range signers {
		pubkey := ss.PublicKey()

		// get wallet state from the network
		walletState, rerr, err := request.GetWalletState(ctx, conn, pubkey)
		if err != nil {
			return nil, err
		}
		if rerr != nil {
			return nil, rerr.Err()
		}
		emitter := false
		for _, t := range walletState.Tags {
			if t == mint.WalletTagEmission.String() {
				emitter = true
				break
			}
		}
		approver := false
		for _, t := range walletState.Tags {
			if t == mint.WalletTagAuthority.String() {
				approver = true
				break
			}
		}

		// check db for a greater nonce
		dbnonce, err := dao.LatestSenderNonce(pubkey)
		if err != nil {
			return nil, err
		}

		// use the greatest
		nonce := walletState.LastTransactionID
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
			approver:    approver,
			signedCount: 0,
		}

		logGold, logMnt := strconv.FormatFloat(walletState.Balance.Gold.Float64(), 'f', 6, 64), strconv.FormatFloat(walletState.Balance.Mnt.Float64(), 'f', 6, 64)
		if emitter {
			logGold, logMnt = "emitter", "emitter"
		}
		logger.
			WithField("net_nonce", walletState.LastTransactionID).
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
