package llo

import (
	"context"
	"fmt"

	"github.com/smartcontractkit/chainlink-common/pkg/services"
	commontypes "github.com/smartcontractkit/chainlink-common/pkg/types"
	"github.com/smartcontractkit/chainlink-data-streams/llo"
	ocrcommontypes "github.com/smartcontractkit/libocr/commontypes"
	ocr2types "github.com/smartcontractkit/libocr/offchainreporting2/types"
	ocr2plus "github.com/smartcontractkit/libocr/offchainreporting2plus"
	ocr3types "github.com/smartcontractkit/libocr/offchainreporting2plus/ocr3types"

	"github.com/smartcontractkit/chainlink/v2/core/logger"
	"github.com/smartcontractkit/chainlink/v2/core/services/job"
	"github.com/smartcontractkit/chainlink/v2/core/services/llo/evm"
	"github.com/smartcontractkit/chainlink/v2/core/services/pg"
	"github.com/smartcontractkit/chainlink/v2/core/services/streams"
)

var _ job.ServiceCtx = &delegate{}

type Closer interface {
	Close() error
}

type delegate struct {
	services.StateMachine

	cfg    DelegateConfig
	codecs map[commontypes.LLOReportFormat]llo.ReportCodec

	llo Closer
}

type DelegateConfig struct {
	Logger   logger.Logger
	Queryer  pg.Queryer
	Runner   streams.Runner
	Registry Registry

	// LLO
	ChannelDefinitionCache commontypes.ChannelDefinitionCache

	// OCR3
	BinaryNetworkEndpointFactory ocr2types.BinaryNetworkEndpointFactory
	V2Bootstrappers              []ocrcommontypes.BootstrapperLocator
	ContractConfigTracker        ocr2types.ContractConfigTracker
	ContractTransmitter          ocr3types.ContractTransmitter[commontypes.LLOReportInfo]
	Database                     ocr3types.Database
	OCRLogger                    ocrcommontypes.Logger
	MonitoringEndpoint           ocrcommontypes.MonitoringEndpoint
	OffchainConfigDigester       ocr2types.OffchainConfigDigester
	OffchainKeyring              ocr2types.OffchainKeyring
	OnchainKeyring               ocr3types.OnchainKeyring[commontypes.LLOReportInfo]
	LocalConfig                  ocr2types.LocalConfig
}

func NewDelegate(cfg DelegateConfig) job.ServiceCtx {
	if cfg.Queryer == nil {
		panic("Queryer must not be nil")
	}
	if cfg.Runner == nil {
		panic("Runner must not be nil")
	}
	if cfg.Registry == nil {
		panic("Registry must not be nil")
	}
	codecs := make(map[commontypes.LLOReportFormat]llo.ReportCodec)

	// NOTE: All codecs must be specified here
	codecs["json"] = llo.JSONReportCodec{}
	codecs["evm"] = evm.ReportCodec{}

	return &delegate{services.StateMachine{}, cfg, codecs, nil}
}

func (d *delegate) Start(ctx context.Context) error {
	return d.StartOnce("LLODelegate", func() error {
		// create the oracle from config values
		// TODO: Do these services need starting?
		prrc := llo.NewPredecessorRetirementReportCache()
		src := llo.NewShouldRetireCache()

		ds := NewDataSource(d.cfg.Logger.Named("DataSource"), d.cfg.Registry)

		llo, err := ocr2plus.NewOracle(ocr2plus.OCR3OracleArgs[commontypes.LLOReportInfo]{
			BinaryNetworkEndpointFactory: d.cfg.BinaryNetworkEndpointFactory,
			V2Bootstrappers:              d.cfg.V2Bootstrappers,
			ContractConfigTracker:        d.cfg.ContractConfigTracker,
			ContractTransmitter:          d.cfg.ContractTransmitter,
			Database:                     d.cfg.Database,
			LocalConfig:                  d.cfg.LocalConfig,
			Logger:                       d.cfg.OCRLogger,
			MonitoringEndpoint:           d.cfg.MonitoringEndpoint,
			OffchainConfigDigester:       d.cfg.OffchainConfigDigester,
			OffchainKeyring:              d.cfg.OffchainKeyring,
			OnchainKeyring:               d.cfg.OnchainKeyring,
			ReportingPluginFactory: llo.NewPluginFactory(
				prrc, src, d.cfg.ChannelDefinitionCache, ds, d.cfg.Logger.Named("LLOReportingPlugin"), d.codecs,
			),
		})

		if err != nil {
			return fmt.Errorf("%w: failed to create new OCR oracle", err)
		}

		d.llo = llo

		return llo.Start()
	})
}

func (d *delegate) Close() error {
	return d.StopOnce("LLODelegate", d.llo.Close)
}
