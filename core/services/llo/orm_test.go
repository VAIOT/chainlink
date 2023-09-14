package llo

import (
	"context"
	"fmt"
	"math/rand"
	"testing"

	commontypes "github.com/smartcontractkit/chainlink-common/pkg/types"

	"github.com/smartcontractkit/chainlink/v2/core/internal/testutils"
	"github.com/smartcontractkit/chainlink/v2/core/internal/testutils/pgtest"
	"github.com/smartcontractkit/chainlink/v2/core/logger"
	"github.com/smartcontractkit/chainlink/v2/core/services/pipeline"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockRunner struct{}

func (m *mockRunner) ExecuteRun(ctx context.Context, spec pipeline.Spec, vars pipeline.Vars, l logger.Logger) (run *pipeline.Run, trrs pipeline.TaskRunResults, err error) {
	return
}

func Test_ORM(t *testing.T) {
	db := pgtest.NewSqlxDB(t)
	orm := NewORM(db, testutils.FixtureChainID)
	ctx := testutils.Context(t)

	addr1 := testutils.NewAddress()
	addr2 := testutils.NewAddress()

	t.Run("LoadChannelDefinitions", func(t *testing.T) {
		t.Run("returns zero values if nothing in database", func(t *testing.T) {
			cd, blockNum, err := orm.LoadChannelDefinitions(ctx, addr1)
			require.NoError(t, err)

			assert.Zero(t, cd)
			assert.Zero(t, blockNum)

		})
		t.Run("loads channel definitions from database", func(t *testing.T) {
			expectedBlockNum := rand.Int63()
			expectedBlockNum2 := rand.Int63()
			cid1 := rand.Uint32()
			cid2 := rand.Uint32()

			channelDefsJSON := fmt.Sprintf(`
{
	"%d": {
		"reportFormat": "example-llo-report-format",
		"chainSelector": 142,
		"streamIds": [1, 2]
	},
	"%d": {
		"reportFormat": "example-llo-report-format",
		"chainSelector": 142,
		"streamIds": [1, 3]
	}
}
			`, cid1, cid2)
			pgtest.MustExec(t, db, `
			INSERT INTO channel_definitions(addr, evm_chain_id, definitions, block_num, created_at, updated_at)
			VALUES (
				$1,
				$2,
				$3,
				$4,
				NOW(),
				NOW()
			)
			`, addr1, testutils.FixtureChainID.String(), channelDefsJSON, expectedBlockNum)

			pgtest.MustExec(t, db, `
			INSERT INTO channel_definitions(addr, evm_chain_id, definitions, block_num, created_at, updated_at)
			VALUES (
				$1,
				$2,
				$3,
				$4,
				NOW(),
				NOW()
			)
			`, addr2, testutils.FixtureChainID.String(), `{}`, expectedBlockNum2)

			cd, blockNum, err := orm.LoadChannelDefinitions(ctx, addr1)
			require.NoError(t, err)

			assert.Equal(t, commontypes.ChannelDefinitions{
				commontypes.ChannelID(cid1): commontypes.ChannelDefinition{
					ReportFormat:  commontypes.LLOReportFormat("example-llo-report-format"),
					ChainSelector: 142,
					StreamIDs:     []commontypes.StreamID{1, 2},
				},
				commontypes.ChannelID(cid2): commontypes.ChannelDefinition{
					ReportFormat:  commontypes.LLOReportFormat("example-llo-report-format"),
					ChainSelector: 142,
					StreamIDs:     []commontypes.StreamID{1, 3},
				},
			}, cd)
			assert.Equal(t, expectedBlockNum, blockNum)

			cd, blockNum, err = orm.LoadChannelDefinitions(ctx, addr2)
			require.NoError(t, err)

			assert.Equal(t, commontypes.ChannelDefinitions{}, cd)
			assert.Equal(t, expectedBlockNum2, blockNum)
		})
	})
}
