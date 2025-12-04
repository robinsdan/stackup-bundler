package execution

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/laizy/log"
	"github.com/robinsdan/sweet"
	"github.com/stackup-wallet/stackup-bundler/pkg/entrypoint"
	"github.com/stackup-wallet/stackup-bundler/pkg/entrypoint/reverts"
	"github.com/stackup-wallet/stackup-bundler/pkg/errors"
	"github.com/stackup-wallet/stackup-bundler/pkg/userop"
)

type SimulateInput struct {
	Rpc        *rpc.Client
	EntryPoint common.Address
	Op         *userop.UserOperation

	// Optional params for simulateHandleOps
	Target common.Address
	Data   []byte
}

type Client struct {
	*ethclient.Client
}

func (self *Client) CallContract(ctx context.Context, msg ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
	log.Infof("calling contract, call msg: %v", sweet.JsonStr(msg))
	val, err := self.Client.CallContract(ctx, msg, blockNumber)
	log.Infof("call result, val:%x, err: %v", val, err)

	return val, err
}

func SimulateHandleOp(in *SimulateInput) (*reverts.ExecutionResultRevert, error) {
	client := ethclient.NewClient(in.Rpc)
	ep, err := entrypoint.NewEntrypoint(in.EntryPoint, &Client{client})
	if err != nil {
		return nil, err
	}

	rawCaller := &entrypoint.EntrypointRaw{Contract: ep}
	err = rawCaller.Call(
		nil,
		nil,
		"simulateHandleOp",
		entrypoint.UserOperation(*in.Op),
		in.Target,
		in.Data,
	)

	sim, simErr := reverts.NewExecutionResult(err)
	if simErr != nil {
		entrypoint.Logger.Error(simErr, "simulate handle op error", "user op", sweet.JsonStr(in.Op))
		fo, foErr := reverts.NewFailedOp(err)
		if foErr != nil {
			if err != nil {
				return nil, err
			}
			return nil, fmt.Errorf("%s, %s", simErr, foErr)
		}
		return nil, errors.NewRPCError(errors.REJECTED_BY_EP_OR_ACCOUNT, fo.Reason, fo)
	}

	return sim, nil
}
