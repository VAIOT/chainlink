package evm

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	commontypes "github.com/smartcontractkit/chainlink-common/pkg/types"
	"github.com/smartcontractkit/chainlink-data-streams/llo"
	"github.com/smartcontractkit/libocr/offchainreporting2plus/types"
)

const ethMainnetChainSelector uint64 = 5009297550715157269

func newValidReport() llo.Report {
	return llo.Report{
		ConfigDigest:      types.ConfigDigest{1, 2, 3},
		ChainSelector:     ethMainnetChainSelector, //
		SeqNr:             32,
		ChannelID:         commontypes.ChannelID(31),
		ValidAfterSeconds: 33,
		ValidUntilSeconds: 34,
		Values:            []*big.Int{big.NewInt(35), big.NewInt(36)},
		Specimen:          true,
	}
}

func Test_ReportCodec(t *testing.T) {
	rc := ReportCodec{}

	t.Run("Encode errors on zero fields", func(t *testing.T) {
		_, err := rc.Encode(llo.Report{})
		require.Error(t, err)

		assert.Contains(t, err.Error(), "failed to get chain ID for selector 0; chain not found for chain selector 0")
	})

	t.Run("Encode constructs a report from observations", func(t *testing.T) {
		report := newValidReport()

		encoded, err := rc.Encode(report)
		require.NoError(t, err)

		reportElems := make(map[string]interface{})
		err = Schema.UnpackIntoMap(reportElems, encoded)
		require.NoError(t, err)

		assert.Equal(t, [32]uint8([32]uint8{0x1, 0x2, 0x3, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}), reportElems["configDigest"])
		assert.Equal(t, uint64(1), reportElems["chainId"])
		assert.Equal(t, uint64(32), reportElems["seqNr"])
		assert.Equal(t, uint32(31), reportElems["channelId"])
		assert.Equal(t, uint32(33), reportElems["validAfterSeconds"])
		assert.Equal(t, uint32(34), reportElems["validUntilSeconds"])
		assert.Equal(t, []*big.Int([]*big.Int{big.NewInt(35), big.NewInt(36)}), reportElems["values"])
		assert.Equal(t, true, reportElems["specimen"])

		assert.Len(t, encoded, 352)
		assert.Equal(t, []byte{0x1, 0x2, 0x3, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x20, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1f, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x21, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x22, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x2, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x23, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x24}, encoded)

		t.Run("Decode decodes the report", func(t *testing.T) {
			decoded, err := rc.Decode(encoded)
			require.NoError(t, err)

			assert.Equal(t, report.ConfigDigest, decoded.ConfigDigest)
			assert.Equal(t, report.ChainSelector, decoded.ChainSelector)
			assert.Equal(t, report.SeqNr, decoded.SeqNr)
			assert.Equal(t, report.ChannelID, decoded.ChannelID)
			assert.Equal(t, report.ValidAfterSeconds, decoded.ValidAfterSeconds)
			assert.Equal(t, report.ValidUntilSeconds, decoded.ValidUntilSeconds)
			assert.Equal(t, report.Values, decoded.Values)
			assert.Equal(t, report.Specimen, decoded.Specimen)
		})
	})

	t.Run("Decode errors on invalid report", func(t *testing.T) {
		_, err := rc.Decode([]byte{1, 2, 3})
		assert.EqualError(t, err, "failed to decode report: abi: cannot marshal in to go type: length insufficient 3 require 32")

		longBad := make([]byte, 64)
		for i := 0; i < len(longBad); i++ {
			longBad[i] = byte(i)
		}
		_, err = rc.Decode(longBad)
		assert.EqualError(t, err, "failed to decode report: abi: improperly encoded uint64 value")
	})
}
