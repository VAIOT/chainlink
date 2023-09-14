package bm

import (
	"context"
	"encoding/json"

	"github.com/smartcontractkit/chainlink-common/pkg/services"
	commontypes "github.com/smartcontractkit/chainlink-common/pkg/types"
	"github.com/smartcontractkit/chainlink/v2/core/logger"
)

// A CDC that loads a static JSON of channel definitions; useful for
// benchmarking and testing

var _ commontypes.ChannelDefinitionCache = &staticCDC{}

type staticCDC struct {
	services.StateMachine
	lggr logger.Logger

	definitions commontypes.ChannelDefinitions
}

func NewStaticChannelDefinitionCache(lggr logger.Logger, dfnstr string) (commontypes.ChannelDefinitionCache, error) {
	var definitions commontypes.ChannelDefinitions
	if err := json.Unmarshal([]byte(dfnstr), &definitions); err != nil {
		return nil, err
	}
	return &staticCDC{services.StateMachine{}, lggr.Named("StaticChannelDefinitionCache"), definitions}, nil
}

func (s *staticCDC) Start(context.Context) error {
	return s.StartOnce("StaticChannelDefinitionCache", func() error {
		return nil
	})
}

func (s *staticCDC) Close() error {
	return s.StopOnce("StaticChannelDefinitionCache", func() error {
		return nil
	})
}

func (s *staticCDC) Definitions() commontypes.ChannelDefinitions {
	return s.definitions
}

func (s *staticCDC) HealthReport() map[string]error {
	report := map[string]error{s.Name(): s.Healthy()}
	return report
}

func (s *staticCDC) Name() string {
	return s.lggr.Name()
}
