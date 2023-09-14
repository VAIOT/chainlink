package llo

import (
	"fmt"

	commontypes "github.com/smartcontractkit/chainlink-common/pkg/types"
	"github.com/smartcontractkit/chainlink-data-streams/llo"
	"github.com/smartcontractkit/chainlink/v2/core/logger"
	"github.com/smartcontractkit/libocr/offchainreporting2/types"
	"github.com/smartcontractkit/libocr/offchainreporting2plus/ocr3types"
	ocrtypes "github.com/smartcontractkit/libocr/offchainreporting2plus/types"
)

type LLOOnchainKeyring ocr3types.OnchainKeyring[commontypes.LLOReportInfo]

var _ LLOOnchainKeyring = &onchainKeyring{}

type Key interface {
	Sign3(digest ocrtypes.ConfigDigest, seqNr uint64, r ocrtypes.Report) (signature []byte, err error)
	Verify3(publicKey ocrtypes.OnchainPublicKey, cd ocrtypes.ConfigDigest, seqNr uint64, r ocrtypes.Report, signature []byte) bool
	PublicKey() ocrtypes.OnchainPublicKey
	MaxSignatureLength() int
}

type onchainKeyring struct {
	evm  Key
	lggr logger.Logger
}

func NewOnchainKeyring(evm Key, lggr logger.Logger) LLOOnchainKeyring {
	return &onchainKeyring{
		evm, lggr.Named("OnchainKeyring"),
	}
}

func (okr *onchainKeyring) PublicKey() types.OnchainPublicKey {
	// TODO: Combine this in some way for multiple chains
	return okr.evm.PublicKey()
}

func (okr *onchainKeyring) MaxSignatureLength() int {
	// TODO: Needs to be max of all chain sigs
	return okr.evm.MaxSignatureLength()
}

func (okr *onchainKeyring) Sign(digest types.ConfigDigest, seqNr uint64, r ocr3types.ReportWithInfo[commontypes.LLOReportInfo]) (signature []byte, err error) {
	switch r.Info.ReportFormat {
	case llo.ReportFormatJSON:
		// FIXME: json signing?
		fallthrough
	case llo.ReportFormatEVM:
		return okr.evm.Sign3(digest, seqNr, r.Report)
	default:
		return nil, fmt.Errorf("Sign failed; unsupported report format: %q", r.Info.ReportFormat)
	}
}

func (okr *onchainKeyring) Verify(key types.OnchainPublicKey, digest types.ConfigDigest, seqNr uint64, r ocr3types.ReportWithInfo[commontypes.LLOReportInfo], signature []byte) bool {
	switch r.Info.ReportFormat {
	case llo.ReportFormatJSON:
		// FIXME: json signing?
		fallthrough
	case llo.ReportFormatEVM:
		return okr.evm.Verify3(key, digest, seqNr, r.Report, signature)
	default:
		okr.lggr.Errorf("cannot verify unrecognized format: %s", r.Info.ReportFormat)
		return false
	}
}
