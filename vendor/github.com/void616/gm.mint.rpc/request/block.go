package request

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"math/big"

	mint "github.com/void616/gm.mint"
	"github.com/void616/gm.mint.rpc/conn"
	"github.com/void616/gm.mint.rpc/rpc"
	"github.com/void616/gm.mint/block"
)

var errHeaderOnly = errors.New("read header only")

// Block model
type Block []byte

// Header parses and returns block's header
func (b Block) Header() (*block.Header, error) {
	var header *block.Header
	var headerCbk = func(h *block.Header) error {
		header = h
		return errHeaderOnly
	}

	// parse block
	if err := block.Parse(bytes.NewBuffer(b), headerCbk, nil); err != nil && err != errHeaderOnly {
		return nil, err
	}
	return header, nil
}

// GetBlockByID method
func GetBlockByID(ctx context.Context, c *conn.Conn, id *big.Int) (res Block, rerr *rpc.Error, err error) {
	res, rerr, err = Block{}, nil, nil

	rctx, rcancel := c.Receive(ctx)
	defer rcancel()

	params := struct {
		ID string `json:"id"`
	}{
		id.String(),
	}

	msg, err := c.Request(rctx, "get_block", params)
	if err != nil {
		return
	}

	switch m := msg.(type) {
	case *rpc.Error:
		rerr = m
		return
	case *rpc.Result:
		str := ""
		err = m.Parse(&str)
		if err != nil {
			return
		}
		var bbytes []byte
		bbytes, err = hex.DecodeString(str)
		if err != nil {
			return
		}
		res = bbytes
		return
	}
	return
}

// GetBlockByDigest method
func GetBlockByDigest(ctx context.Context, c *conn.Conn, digest mint.Digest) (res Block, rerr *rpc.Error, err error) {
	res, rerr, err = Block{}, nil, nil

	rctx, rcancel := c.Receive(ctx)
	defer rcancel()

	params := struct {
		Digest string `json:"digest"`
	}{
		digest.String(),
	}

	msg, err := c.Request(rctx, "get_block", params)
	if err != nil {
		return
	}

	switch m := msg.(type) {
	case *rpc.Error:
		rerr = m
		return
	case *rpc.Result:
		str := ""
		err = m.Parse(&str)
		if err != nil {
			return
		}
		var bbytes []byte
		bbytes, err = hex.DecodeString(str)
		if err != nil {
			return
		}
		res = bbytes
		return
	}
	return
}
