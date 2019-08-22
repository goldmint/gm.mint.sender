package fee

import (
	"math/big"

	"github.com/void616/gm-sumuslib/amount"
)

var (
	goldMinFixed     = amount.NewFloatString("0.00002")
	goldMaxFixed     = amount.NewFloatString("0.002")
	mntFixed         = amount.NewFloatString("0.02")
	mntGradient10    = amount.NewFloatString("10")
	mntGradient1000  = amount.NewFloatString("1000")
	mntGradient10000 = amount.NewFloatString("10000")
	mntPerByte       = amount.NewFloatString("0.004")
)

// GoldFee calculation
func GoldFee(g *amount.Amount, m *amount.Amount) *amount.Amount {
	ret := new(big.Int).Set(g.Value)

	switch {
	// >= 10k
	case m.Value.Cmp(mntGradient10000.Value) >= 0:
		ret.Mul(ret, big.NewInt(3))
		//ret.Div(ret, big.NewInt(100000))
		divRounding(ret, 100000)
		if ret.Cmp(goldMaxFixed.Value) > 0 {
			ret.Set(goldMaxFixed.Value)
		}
	// >= 1k
	case m.Value.Cmp(mntGradient1000.Value) >= 0:
		ret.Mul(ret, big.NewInt(3))
		//ret.Div(ret, big.NewInt(100000))
		divRounding(ret, 100000)
	// >= 10
	case m.Value.Cmp(mntGradient10.Value) >= 0:
		ret.Mul(ret, big.NewInt(3))
		//ret.Div(ret, big.NewInt(10000))
		divRounding(ret, 10000)
	// < 10
	default:
		ret.Mul(ret, big.NewInt(1))
		//ret.Div(ret, big.NewInt(1000))
		divRounding(ret, 1000)
	}

	// minimal
	if ret.Cmp(goldMinFixed.Value) < 0 {
		ret.Set(goldMinFixed.Value)
	}

	return amount.NewBig(ret)
}

// MntFee calculation
func MntFee(m *amount.Amount) *amount.Amount {
	return amount.NewAmount(mntFixed)
}

// UserDataFee calculation
func UserDataFee(l uint32) *amount.Amount {
	ret := big.NewInt(int64(l))
	ret.Mul(ret, mntPerByte.Value)
	return amount.NewBig(ret)
}

// ---

func divRounding(x *big.Int, y int64) {
	ten := big.NewInt(10)
	x.Mul(x, ten)
	x.Div(x, big.NewInt(y))
	m := new(big.Int).Mod(x, ten)
	x.Div(x, ten)
	if m.Cmp(big.NewInt(5)) >= 0 {
		x.Add(x, big.NewInt(1))
		return
	}
}
