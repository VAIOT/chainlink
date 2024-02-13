package llo

import (
	"sync"

	"github.com/ethereum/go-ethereum/common"

	commontypes "github.com/smartcontractkit/chainlink-common/pkg/types"

	"github.com/smartcontractkit/chainlink/v2/core/chains/evm/logpoller"
	"github.com/smartcontractkit/chainlink/v2/core/logger"
	lloconfig "github.com/smartcontractkit/chainlink/v2/core/services/ocr2/plugins/llo/config"
)

type ChannelDefinitionCacheFactory interface {
	NewCache(cfg lloconfig.PluginConfig) (commontypes.ChannelDefinitionCache, error)
}

var _ ChannelDefinitionCacheFactory = &channelDefinitionCacheFactory{}

func NewChannelDefinitionCacheFactory(lggr logger.Logger, orm ChannelDefinitionCacheORM, lp logpoller.LogPoller) ChannelDefinitionCacheFactory {
	return &channelDefinitionCacheFactory{
		lggr,
		orm,
		lp,
		make(map[common.Address]struct{}),
		sync.Mutex{},
	}
}

type channelDefinitionCacheFactory struct {
	lggr logger.Logger
	orm  ChannelDefinitionCacheORM // TODO: pass in a pre-scoped ORM (to EVM chain ID)
	lp   logpoller.LogPoller

	caches map[common.Address]struct{}
	mu     sync.Mutex
}

func (f *channelDefinitionCacheFactory) NewCache(cfg lloconfig.PluginConfig) (commontypes.ChannelDefinitionCache, error) {
	if cfg.ChannelDefinitions != "" {
		return NewStaticChannelDefinitionCache(f.lggr, cfg.ChannelDefinitions)
	}

	addr := cfg.ChannelDefinitionsContractAddress
	fromBlock := cfg.ChannelDefinitionsContractFromBlock

	f.mu.Lock()
	defer f.mu.Unlock()

	if _, exists := f.caches[addr]; exists {
		// TODO: can we do better?
		panic("cannot create duplicate cache")
	}
	f.caches[addr] = struct{}{}
	return NewChannelDefinitionCache(f.lggr, f.orm, f.lp, addr, fromBlock), nil
}
