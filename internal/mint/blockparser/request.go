package blockparser

import (
	"bytes"
	"io"
	"io/ioutil"
	"math/big"

	"github.com/void616/gm.mint.rpc/request"
)

// queryBlockData queries block data via RPC connection
func (p *Parser) queryBlockData(block *big.Int) (io.ReadCloser, error) {

	// get connection
	ctx, conn, cls, err := p.rpcpool.Conn()
	if err != nil {
		return nil, err
	}
	defer cls()

	// get the block
	blockData, rerr, err := request.GetBlockByID(ctx, conn, block)
	if rerr != nil {
		return nil, rerr.Err()
	}
	if err != nil {
		return nil, err
	}

	blockStream := bytes.NewBuffer(blockData[:])
	return ioutil.NopCloser(blockStream), nil
}
