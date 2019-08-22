package mysql

import (
	"bytes"
	"encoding/json"
	"math/big"
	"testing"
	"time"

	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/void616/gm-mint-sender/internal/watcher/db"
	"github.com/void616/gm-mint-sender/internal/watcher/db/types"
	sumuslib "github.com/void616/gm-sumuslib"
	"github.com/void616/gm-sumuslib/amount"
	"github.com/void616/gm-sumuslib/signer"
	gormigrate "gopkg.in/gormigrate.v1"
)

var (
	tablepfx = "wtc_"
	database = "mint_watcher_test"
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
		pub3 sumuslib.PublicKey
	)
	{
		s1, _ := signer.New()
		pub1 = s1.PublicKey()
		s2, _ := signer.New()
		pub2 = s2.PublicKey()
		s3, _ := signer.New()
		pub3 = s3.PublicKey()
	}

	// no wallets now
	wallets, err := dao.ListWallets()
	if err != nil {
		t.Fatal(err)
	}
	if len(wallets) > 0 {
		t.Fatal("not empty")
	}

	// add 2 wallets
	if err := dao.PutWallet(&types.PutWallet{
		PublicKeys: []sumuslib.PublicKey{
			pub1, pub2,
		},
	}); err != nil {
		t.Fatal(err)
	}

	// check 2 wallets added
	wallets, err = dao.ListWallets()
	if err != nil {
		t.Fatal(err)
	}
	if len(wallets) != 2 {
		t.Fatal("not added")
	}
	for _, w := range wallets {
		if !bytes.Equal(w.PublicKey[:], pub1[:]) && !bytes.Equal(w.PublicKey[:], pub2[:]) {
			t.Fatal("wrong pubkey bytes")
		}
	}

	// add 3 wallets (2 dups)
	if err := dao.PutWallet(&types.PutWallet{
		PublicKeys: []sumuslib.PublicKey{
			pub1, pub2, pub3,
		},
	}); err != nil {
		t.Fatal(err)
	}

	// 1 added
	wallets, err = dao.ListWallets()
	if err != nil {
		t.Fatal(err)
	}
	if len(wallets) != 3 {
		t.Fatal("not added")
	}

	// remove 2 wallets
	if err := dao.DeleteWallet(&types.DeleteWallet{
		PublicKeys: []sumuslib.PublicKey{
			pub1, pub2,
		},
	}); err != nil {
		t.Fatal(err)
	}

	// 1 remaining
	wallets, err = dao.ListWallets()
	if err != nil {
		t.Fatal(err)
	}
	if len(wallets) != 1 {
		t.Fatal("not added")
	}
	if !bytes.Equal(wallets[0].PublicKey[:], pub3[:]) && !bytes.Equal(wallets[0].PublicKey[:], pub3[:]) {
		t.Fatal("wrong remaining")
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
	s3, _ := signer.New()

	// incoming
	var inc1 *types.PutIncoming
	var inc2 *types.PutIncoming
	var inc3 *types.PutIncoming
	{
		// truncate nano
		time := time.Unix(time.Now().UTC().Unix(), 0).UTC()

		// s1 -> 666 GOLD -> s2
		inc1 = &types.PutIncoming{
			From:      s1.PublicKey(),
			To:        s2.PublicKey(),
			Amount:    amount.NewFloatString("666.123456789123456789"),
			Token:     sumuslib.TokenGOLD,
			Digest:    sumuslib.Digest(s3.PublicKey()),
			Block:     big.NewInt(1),
			Timestamp: time,
		}
		// s2 -> 0.00...1 MNT -> s3
		inc2 = &types.PutIncoming{
			From:      s2.PublicKey(),
			To:        s3.PublicKey(),
			Amount:    amount.NewFloatString("0.000000000000000001"),
			Token:     sumuslib.TokenMNT,
			Digest:    sumuslib.Digest(s1.PublicKey()),
			Block:     big.NewInt(2),
			Timestamp: time,
		}
		// s3 -> 1 GOLD -> s1
		inc3 = &types.PutIncoming{
			From:      s3.PublicKey(),
			To:        s1.PublicKey(),
			Amount:    amount.NewFloatString("1"),
			Token:     sumuslib.TokenGOLD,
			Digest:    sumuslib.Digest(s2.PublicKey()),
			Block:     big.NewInt(3),
			Timestamp: time,
		}
	}

	// add incomings
	if err := dao.PutIncoming(inc1); err != nil {
		t.Fatal(err)
	}
	if err := dao.PutIncoming(inc2); err != nil {
		t.Fatal(err)
	}
	if err := dao.PutIncoming(inc3); err != nil {
		t.Fatal(err)
	}

	check := func(inc *types.ListUnsentIncomingsItem) bool {
		for _, c := range []*types.PutIncoming{inc1, inc2, inc3} {
			if bytes.Equal(inc.Digest[:], c.Digest[:]) {
				if bytes.Equal(inc.From[:], c.From[:]) &&
					bytes.Equal(inc.To[:], c.To[:]) &&
					inc.Amount.String() == c.Amount.String() &&
					inc.Token == c.Token &&
					inc.Block.String() == c.Block.String() &&
					inc.Timestamp == c.Timestamp {
					return true
				}
				j1, _ := json.Marshal(c)
				j2, _ := json.Marshal(inc)
				t.Log(string(j1), "!=", string(j2))
			}
		}
		return false
	}

	// get 3 unsent incomings
	incomings, err := dao.ListUnsentIncomings(0xFFFF)
	if err != nil {
		t.Fatal(err)
	}
	if len(incomings) != 3 {
		t.Fatal("not added")
	}
	for _, inc := range incomings {
		if !check(inc) {
			t.Fatal("saving or mapping problems")
		}
	}

	// get max 2 unsent incomings
	incomings, err = dao.ListUnsentIncomings(2)
	if err != nil {
		t.Fatal(err)
	}
	if len(incomings) != 2 {
		t.Fatal("limit doesnt work")
	}
	// and mark em as sent
	for _, inc := range incomings {
		if err := dao.MarkIncomingSent(&types.MarkIncomingSent{
			Digest: inc.Digest,
			Sent:   true,
		}); err != nil {
			t.Fatal(err)
		}
	}

	// add dups
	if err := dao.PutIncoming(inc1); err != nil {
		t.Fatal(err)
	}
	if err := dao.PutIncoming(inc2); err != nil {
		t.Fatal(err)
	}
	if err := dao.PutIncoming(inc3); err != nil {
		t.Fatal(err)
	}

	// only 1 unsent should left now
	incomings, err = dao.ListUnsentIncomings(0xFFFF)
	if err != nil {
		t.Fatal(err)
	}
	if len(incomings) != 1 {
		t.Fatal("mark as sent failed")
	}
}
