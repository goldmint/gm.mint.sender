package mysql

import (
	"bytes"
	"math/big"
	"testing"

	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/void616/gm-mint-sender/internal/sender/db"
	sumuslib "github.com/void616/gm-sumuslib"
	"github.com/void616/gm-sumuslib/amount"
	"github.com/void616/gm-sumuslib/signer"
	gormigrate "gopkg.in/gormigrate.v1"
)

var (
	tablepfx = "snd_"
	database = "mint_sender_test"
	dsn      = "root:000000@tcp(localhost:3306)/" + database + "?collation=utf8_general_ci&timeout=10s&readTimeout=60s&writeTimeout=60s"
)

func TestWallets(t *testing.T) {

	var dao db.DAO

	{
		db, err := New(dsn, tablepfx, false, 16*1024*1024)
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()

		db.DropTableIfExists(AutoMigration...)
		db.DropTableIfExists(tablepfx + "dbmigrations")

		opts := gormigrate.DefaultOptions
		opts.TableName = tablepfx + "dbmigrations"
		mig := gormigrate.New(db.DB, opts, Migrations)
		if err := mig.Migrate(); err != nil {
			t.Fatal(err)
		}

		dao = db
		if !dao.Available() {
			t.Fatal(err)
		}
	}

	// wallets
	var (
		pub1 sumuslib.PublicKey
		pub2 sumuslib.PublicKey
	)
	{
		s1, _ := signer.New()
		pub1 = s1.PublicKey()
		s2, _ := signer.New()
		pub2 = s2.PublicKey()
	}

	// no wallets now
	wallets, err := dao.ListSenderWallets()
	if err != nil {
		t.Fatal(err)
	}
	if len(wallets) > 0 {
		t.Fatal("not empty")
	}

	// add 2 wallets
	if err := dao.SaveSenderWallet(pub1); err != nil {
		t.Fatal(err)
	}
	if err := dao.SaveSenderWallet(pub1); err != nil {
		t.Fatal(err)
	}
	if err := dao.SaveSenderWallet(pub2); err != nil {
		t.Fatal(err)
	}

	// check 2 wallets added
	wallets, err = dao.ListSenderWallets()
	if err != nil {
		t.Fatal(err)
	}
	if len(wallets) != 2 {
		t.Fatal("not added")
	}
	for _, w := range wallets {
		if !bytes.Equal(w[:], pub1[:]) && !bytes.Equal(w[:], pub2[:]) {
			t.Fatal("wrong pubkey bytes")
		}
	}
}

func TestIncomings(t *testing.T) {

	var dao db.DAO

	{
		db, err := New(dsn, tablepfx, false, 16*1024*1024)
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()

		db.DropTableIfExists(AutoMigration...)
		db.DropTableIfExists(tablepfx + "dbmigrations")

		opts := gormigrate.DefaultOptions
		opts.TableName = tablepfx + "dbmigrations"
		mig := gormigrate.New(db.DB, opts, Migrations)
		if err := mig.Migrate(); err != nil {
			t.Fatal(err)
		}

		dao = db
		if !dao.Available() {
			t.Fatal(err)
		}
	}

	// signers
	s1, _ := signer.New()
	s2, _ := signer.New()

	type Sending struct {
		To     sumuslib.PublicKey
		Amount string
		Token  sumuslib.Token
		Digest sumuslib.Digest
	}

	snd1 := Sending{
		To:     s2.PublicKey(),
		Amount: "1.234000000000000000",
		Token:  sumuslib.TokenGOLD,
		Digest: sumuslib.Digest(s1.PublicKey()),
	}
	snd2 := Sending{
		To:     s1.PublicKey(),
		Amount: "0.567000000000000008",
		Token:  sumuslib.TokenMNT,
		Digest: sumuslib.Digest(s2.PublicKey()),
	}
	snds := []Sending{snd1, snd2}

	// enqueue sendings
	if err := dao.EnqueueSending(
		"snd1",
		snd1.To,
		amount.NewFloatString(snd1.Amount),
		snd1.Token,
	); err != nil {
		t.Fatal(err)
	}
	if err := dao.EnqueueSending(
		"snd2",
		snd2.To,
		amount.NewFloatString(snd2.Amount),
		snd2.Token,
	); err != nil {
		t.Fatal(err)
	}
	if err := dao.EnqueueSending(
		"snd1",
		snd1.To,
		amount.NewFloatString(snd1.Amount),
		snd1.Token,
	); err == nil {
		t.Fatal("should fail")
	}

	// empty earliest block
	if eb, err := dao.EarliestBlock(); err != nil {
		t.Fatal(err)
	} else if !eb.Empty {
		t.Fatal("not empty")
	}

	// zero latest nonce for signer 1
	if ln, err := dao.LatestSenderNonce(s1.PublicKey()); err != nil || ln != 0 {
		t.Fatal("nonce", err)
	}
	// zero latest nonce for signer 2
	if ln, err := dao.LatestSenderNonce(s2.PublicKey()); err != nil || ln != 0 {
		t.Fatal("nonce", err)
	}

	// limited enqueued list
	if list, err := dao.ListEnqueuedSendings(1); err != nil || len(list) != 1 {
		t.Fatal(err)
	}

	// 2 sendings enqueued
	enqList, err := dao.ListEnqueuedSendings(0xFFFF)
	if err != nil || len(enqList) != 2 {
		t.Fatal(err)
	}
	for i, v := range enqList {
		if v.Token != snds[i].Token || !bytes.Equal(v.To[:], snds[i].To[:]) || v.Amount.String() != snds[i].Amount {
			t.Fatal("mapping")
		}
	}

	// ---

	latestBlock := big.NewInt(10)

	// 1st is posted
	if err := dao.SetSendingPosted(
		1,
		s1.PublicKey(),
		1,
		snd1.Digest,
		latestBlock,
	); err != nil {
		t.Fatal(err)
	}

	// latest nonce for signer 1
	if ln, err := dao.LatestSenderNonce(s1.PublicKey()); err != nil || ln != 1 {
		t.Fatal("nonce", err)
	}
	// zero latest nonce for signer 2
	if ln, err := dao.LatestSenderNonce(s2.PublicKey()); err != nil || ln != 0 {
		t.Fatal("nonce", err)
	}

	// earliest block now is 1
	if eb, err := dao.EarliestBlock(); err != nil {
		t.Fatal(err)
	} else if eb.Empty || eb.Block.Cmp(latestBlock) != 0 {
		t.Fatal("wrong earliest block")
	}

	// 1 enqueued in the list now
	if list, err := dao.ListEnqueuedSendings(0xFF); err != nil || len(list) != 1 {
		t.Fatal(err)
	}

	// ---

	// some blocks later...
	latestBlock.SetUint64(20)

	// 2st is posted
	if err := dao.SetSendingPosted(
		2,
		s2.PublicKey(),
		2,
		snd2.Digest,
		latestBlock,
	); err != nil {
		t.Fatal(err)
	}

	// latest nonce for signer 1
	if ln, err := dao.LatestSenderNonce(s1.PublicKey()); err != nil || ln != 1 {
		t.Fatal("nonce", err)
	}
	// latest nonce for signer 2
	if ln, err := dao.LatestSenderNonce(s2.PublicKey()); err != nil || ln != 2 {
		t.Fatal("nonce", err)
	}

	// earliest block now is 1st posted block
	if eb, err := dao.EarliestBlock(); err != nil {
		t.Fatal(err)
	} else if eb.Empty || eb.Block.Cmp(latestBlock) >= 0 {
		t.Fatal("wrong earliest block")
	}

	// 0 enqueued in the list now
	if list, err := dao.ListEnqueuedSendings(0xFF); err != nil || len(list) != 0 {
		t.Fatal(err)
	}

	// ---

	// 1st request became stale
	staleList, err := dao.ListStaleSendings(latestBlock, 0xFF)
	if err != nil || len(staleList) != 1 {
		t.Fatal(err)
	}
	if staleList[0].Token != snd1.Token || !bytes.Equal(staleList[0].To[:], snd1.To[:]) || staleList[0].Amount.String() != snd1.Amount {
		t.Fatal("mapping")
	}

	// 1st is reposted
	if err := dao.SetSendingPosted(
		staleList[0].ID,
		s1.PublicKey(),
		1,
		snd1.Digest,
		latestBlock,
	); err != nil {
		t.Fatal(err)
	}

	// 0 stale sendings
	staleList, err = dao.ListStaleSendings(latestBlock, 0xFF)
	if err != nil || len(staleList) != 0 {
		t.Fatal(err)
	}

	// ---

	// some blocks later...
	latestBlock.SetUint64(30)

	// confirm sendings
	if err := dao.SetSendingConfirmed(
		s1.PublicKey(),
		snd1.Digest,
		latestBlock,
	); err != nil {
		t.Fatal(err)
	}
	if err := dao.SetSendingConfirmed(
		s2.PublicKey(),
		snd2.Digest,
		latestBlock,
	); err != nil {
		t.Fatal(err)
	}

	// 0 stale sendings
	staleList, err = dao.ListStaleSendings(latestBlock, 0xFF)
	if err != nil || len(staleList) != 0 {
		t.Fatal(err)
	}

	// 0 enqueued list
	enqList, err = dao.ListEnqueuedSendings(0xFFFF)
	if err != nil || len(enqList) != 0 {
		t.Fatal(err)
	}

	// ---

	// 2 in unnotified list
	unList, err := dao.ListUnnotifiedSendings(0xFFFF)
	if err != nil || len(unList) != 2 {
		t.Fatal(err)
	}

	// limit unnotified list
	unList, err = dao.ListUnnotifiedSendings(1)
	if err != nil || len(unList) != 1 {
		t.Fatal(err)
	}

	// 1st notified
	if err := dao.SetSendingNotified(
		1,
		true,
	); err != nil {
		t.Fatal(err)
	}
	// 1 in unnotified list
	unList, err = dao.ListUnnotifiedSendings(0xFFFF)
	if err != nil || len(unList) != 1 {
		t.Fatal(err)
	}
	if unList[0].RequestID != "snd2" {
		t.Fatal("oups")
	}
	// 2nd
	if err := dao.SetSendingNotified(
		2,
		true,
	); err != nil {
		t.Fatal(err)
	}
	// 0 in unnotified list
	unList, err = dao.ListUnnotifiedSendings(0xFFFF)
	if err != nil || len(unList) != 0 {
		t.Fatal(err)
	}

	// 0 stale sendings
	staleList, err = dao.ListStaleSendings(latestBlock, 0xFF)
	if err != nil || len(staleList) != 0 {
		t.Fatal(err)
	}

	// 0 enqueued list
	enqList, err = dao.ListEnqueuedSendings(0xFFFF)
	if err != nil || len(enqList) != 0 {
		t.Fatal(err)
	}

	// ---

	// 1st confirmed again

	// confirm sendings
	if err := dao.SetSendingConfirmed(
		s1.PublicKey(),
		snd1.Digest,
		latestBlock,
	); err != nil {
		t.Fatal(err)
	}

	// 0 in unnotified list
	unList, err = dao.ListUnnotifiedSendings(0xFFFF)
	if err != nil || len(unList) != 0 {
		t.Fatal(err)
	}

	// 0 stale sendings
	staleList, err = dao.ListStaleSendings(latestBlock, 0xFF)
	if err != nil || len(staleList) != 0 {
		t.Fatal(err)
	}

	// 0 enqueued list
	enqList, err = dao.ListEnqueuedSendings(0xFFFF)
	if err != nil || len(enqList) != 0 {
		t.Fatal(err)
	}
}
