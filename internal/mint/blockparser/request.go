package blockparser

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"math/big"

	"github.com/void616/gm-sumusrpc/rpc"
)

// queryBlockData queries block data via RPC connection
func (p *Parser) queryBlockData(block *big.Int) (io.ReadCloser, error) {

	// get connection
	conn, err := p.rpcpool.Get()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// get the block
	blockString, code, err := rpc.BlockData(conn.Conn(), block)
	if code != rpc.ECSuccess {
		return nil, fmt.Errorf("node error code %v", code)
	}
	if err != nil {
		return nil, err
	}

	// parse
	blockBytes, err := hex.DecodeString(blockString)
	if err != nil {
		return nil, err
	}
	blockStream := bytes.NewBuffer(blockBytes)

	return ioutil.NopCloser(blockStream), nil
}
