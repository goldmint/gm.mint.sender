package nats

/*
import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	proto "github.com/golang/protobuf/proto"
	gonats "github.com/nats-io/go-nats"
	"github.com/sirupsen/logrus"
	walletsvc "github.com/void616/gm-mint-sender/pkg/watcher/nats/wallet"
	sumuslib "github.com/void616/gm-sumuslib"
	"github.com/void616/gotask"
)

const url = "localhost:4222"

var walletInvalid = []byte{0x00, 0x11, 0x22, 0x33}
var wallet1 = []byte{0xDE, 0xAD, 0xBE, 0xEF, 0xDE, 0xAD, 0xBE, 0xEF, 0xDE, 0xAD, 0xBE, 0xEF, 0xDE, 0xAD, 0xBE, 0xEF, 0xDE, 0xAD, 0xBE, 0xEF, 0xDE, 0xAD, 0xBE, 0xEF, 0xDE, 0xAD, 0xBE, 0xEF, 0xDE, 0xAD, 0xBE, 0xEF}
var wallet2 = []byte{0x13, 0x37, 0x13, 0x37, 0x13, 0x37, 0x13, 0x37, 0x13, 0x37, 0x13, 0x37, 0x13, 0x37, 0x13, 0x37, 0x13, 0x37, 0x13, 0x37, 0x13, 0x37, 0x13, 0x37, 0x13, 0x37, 0x13, 0x37, 0x13, 0x37, 0x13, 0x37}

func TestNew(t *testing.T) {

	log := logrus.New()
	mwsvc := &mockWalletService{Works: true}

	svc, cls, err := New(url, "prefix", mwsvc, logrus.NewEntry(log))
	if err != nil {
		t.Fatal(err)
	}
	defer cls()

	task, _ := gotask.NewTask("test", svc.Task)
	taskt, taskw, _ := task.Run()

	// ---

	{
		nc, err := gonats.Connect(url)
		if err != nil {
			t.Fatal(err)
		}
		wg := sync.WaitGroup{}
		wg.Add(3)

		// add 50 * 2, remove 50 * 2
		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				req, _ := proto.Marshal(&walletsvc.AddRemoveRequest{Add: i%2 == 0, Pubkey: []string{sumuslib.Pack58(wallet1), sumuslib.Pack58(wallet2)}})
				msg, err := nc.Request(svc.subjPrefix+walletsvc.SubjectWatch, req, time.Second*5)
				if err != nil || msg == nil {
					t.Fatal(err)
				}
				rep := walletsvc.AddRemoveReply{}
				if err := proto.Unmarshal(msg.Data, &rep); err != nil {
					t.Fatal(err)
				} else if !rep.Success {
					t.Fatal("fail")
				}
			}
		}()

		// add 50, remove 50
		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				req, _ := proto.Marshal(&walletsvc.AddRemoveRequest{Add: i%2 == 0, Pubkey: []string{sumuslib.Pack58(wallet1)}})
				msg, err := nc.Request(svc.subjPrefix+walletsvc.SubjectWatch, req, time.Second*5)
				if err != nil || msg == nil {
					t.Fatal(err)
				}
				rep := walletsvc.AddRemoveReply{}
				if err := proto.Unmarshal(msg.Data, &rep); err != nil {
					t.Fatal(err)
				} else if !rep.Success {
					t.Fatal("fail")
				}
			}
		}()

		// invalid
		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				req, _ := proto.Marshal(&walletsvc.AddRemoveRequest{Add: i%2 == 0, Pubkey: []string{sumuslib.Pack58(walletInvalid)}})
				msg, err := nc.Request(svc.subjPrefix+walletsvc.SubjectWatch, req, time.Second*5)
				if err != nil || msg == nil {
					t.Fatal(err)
				}
				rep := walletsvc.AddRemoveReply{}
				if err := proto.Unmarshal(msg.Data, &rep); err != nil {
					t.Fatal(err)
				} else if rep.Success {
					t.Fatal("success")
				}
			}
		}()

		wg.Wait()

		mwsvc.Works = false
		wg.Add(1)

		// fails
		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				req, _ := proto.Marshal(&walletsvc.AddRemoveRequest{Add: i%2 == 0, Pubkey: []string{sumuslib.Pack58(wallet1)}})
				msg, err := nc.Request(svc.subjPrefix+walletsvc.SubjectWatch, req, time.Second*5)
				if err != nil || msg == nil {
					t.Fatal(err)
				}
				rep := walletsvc.AddRemoveReply{}
				if err := proto.Unmarshal(msg.Data, &rep); err != nil {
					t.Fatal(err)
				}
				if rep.Success {
					t.Fatal("success")
				}
			}
		}()

		wg.Wait()
	}
	taskt.Stop()

	// ---

	taskw.Wait()
	if mwsvc.Adds != 150 {
		t.Fatal("fail")
	}
	if mwsvc.Rems != 150 {
		t.Fatal("fail")
	}
}

// ---

type mockWalletService struct {
	Works bool
	Adds  uint32
	Rems  uint32
}

func (s *mockWalletService) AddWallet(pubs ...sumuslib.PublicKey) bool {
	if !s.Works {
		return false
	}
	atomic.AddUint32(&s.Adds, uint32(len(pubs)))
	return true
}

func (s *mockWalletService) RemoveWallet(pubs ...sumuslib.PublicKey) bool {
	if !s.Works {
		return false
	}
	atomic.AddUint32(&s.Rems, uint32(len(pubs)))
	return true
}
*/
