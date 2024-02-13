package llo

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	commontypes "github.com/smartcontractkit/chainlink-common/pkg/types"

	"github.com/smartcontractkit/chainlink/v2/core/services/pg"
)

type ORM interface {
	ChannelDefinitionCacheORM
}

var _ ORM = &orm{}

type orm struct {
	q          pg.Queryer
	evmChainID *big.Int
}

func NewORM(q pg.Queryer, evmChainID *big.Int) ORM {
	return &orm{q, evmChainID}
	// TODO: make sure to scope by chain ID everywhere
}

func (o *orm) LoadChannelDefinitions(ctx context.Context, addr common.Address) (cd commontypes.ChannelDefinitions, blockNum int64, err error) {
	type scd struct {
		Definitions []byte `db:"definitions"`
		BlockNum    int64  `db:"block_num"`
	}
	var scanned scd
	err = o.q.GetContext(ctx, &scanned, "SELECT definitions, block_num FROM channel_definitions WHERE evm_chain_id = $1 AND addr = $2", o.evmChainID.String(), addr)
	if errors.Is(err, sql.ErrNoRows) {
		return cd, blockNum, nil
	} else if err != nil {
		return nil, 0, fmt.Errorf("failed to LoadChannelDefinitions; %w", err)
	}

	if err = json.Unmarshal(scanned.Definitions, &cd); err != nil {
		return nil, 0, fmt.Errorf("failed to LoadChannelDefinitions; JSON Unmarshal failure; %w", err)
	}

	return cd, scanned.BlockNum, nil
}

func (o *orm) StoreChannelDefinitions(ctx context.Context, cd commontypes.ChannelDefinitions, blockNum int64) error {
	_, err := o.q.ExecContext(ctx, `
INSERT INTO channel_definitions (evm_chain_id, addr, definitions, block_num)
VALUES ($1, $2, $3, $4)
ON CONFLICT (evm_chain_id, addr) DO UPDATE
SET definitions = $3, block_num = $4
`)
	return fmt.Errorf("StoreChannelDefinitions failed: %w", err)
}
