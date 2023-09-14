package evm

// TODO: Move all evm-related llo code into this package

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"

	chainselectors "github.com/smartcontractkit/chain-selectors"
	commontypes "github.com/smartcontractkit/chainlink-common/pkg/types"
	"github.com/smartcontractkit/chainlink-data-streams/llo"
	"github.com/smartcontractkit/libocr/offchainreporting2/types"
)

var (
	_      llo.ReportCodec = ReportCodec{}
	Schema                 = getSchema()
)

func getSchema() abi.Arguments {
	mustNewType := func(t string) abi.Type {
		result, err := abi.NewType(t, "", []abi.ArgumentMarshaling{})
		if err != nil {
			panic(fmt.Sprintf("Unexpected error during abi.NewType: %s", err))
		}
		return result
	}
	return abi.Arguments([]abi.Argument{
		{Name: "configDigest", Type: mustNewType("bytes32")},
		{Name: "chainId", Type: mustNewType("uint64")},
		// TODO:
		// could also include address of verifier to make things more specific.
		// downside is increased data size.
		// for now we assume that a channelId will only be registered on a single
		// verifier per chain.
		{Name: "seqNr", Type: mustNewType("uint64")},
		{Name: "channelId", Type: mustNewType("uint32")},
		{Name: "validAfterSeconds", Type: mustNewType("uint32")},
		{Name: "validUntilSeconds", Type: mustNewType("uint32")},
		{Name: "values", Type: mustNewType("int192[]")},
		{Name: "specimen", Type: mustNewType("bool")},
	})
}

type ReportCodec struct{}

func NewReportCodec() ReportCodec {
	return ReportCodec{}
}

func (rc ReportCodec) Encode(report llo.Report) (b []byte, merr error) {
	chainID, err := chainselectors.ChainIdFromSelector(report.ChainSelector)
	if err != nil {
		merr = errors.Join(merr, fmt.Errorf("failed to get chain ID for selector %d; %w", report.ChainSelector, err))
	}

	if merr != nil {
		return nil, merr
	}
	b, merr = Schema.Pack(report.ConfigDigest, chainID, report.SeqNr, report.ChannelID, report.ValidAfterSeconds, report.ValidUntilSeconds, report.Values, report.Specimen)
	return
}

func (r ReportCodec) Decode(encoded []byte) (llo.Report, error) {
	type decode struct {
		ConfigDigest      types.ConfigDigest
		ChainId           uint64
		SeqNr             uint64
		ChannelId         commontypes.ChannelID
		ValidAfterSeconds uint32
		ValidUntilSeconds uint32
		Values            []*big.Int
		Specimen          bool
	}
	values, err := Schema.Unpack(encoded)
	if err != nil {
		return llo.Report{}, fmt.Errorf("failed to decode report: %w", err)
	}
	decoded := new(decode)
	if err = Schema.Copy(decoded, values); err != nil {
		return llo.Report{}, fmt.Errorf("failed to copy report values to struct: %w", err)
	}
	chainSelector, err := chainselectors.SelectorFromChainId(decoded.ChainId)
	return llo.Report{
		ConfigDigest:      decoded.ConfigDigest,
		ChainSelector:     chainSelector,
		SeqNr:             decoded.SeqNr,
		ChannelID:         decoded.ChannelId,
		ValidAfterSeconds: decoded.ValidAfterSeconds,
		ValidUntilSeconds: decoded.ValidUntilSeconds,
		Values:            decoded.Values,
		Specimen:          decoded.Specimen,
	}, err
}
