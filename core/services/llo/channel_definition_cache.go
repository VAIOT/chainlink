package llo

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"maps"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/smartcontractkit/chainlink-common/pkg/services"
	commontypes "github.com/smartcontractkit/chainlink-common/pkg/types"

	"github.com/smartcontractkit/chainlink/v2/core/chains/evm/logpoller"
	"github.com/smartcontractkit/chainlink/v2/core/gethwrappers/llo-feeds/generated/channel_config_store"
	"github.com/smartcontractkit/chainlink/v2/core/logger"
	"github.com/smartcontractkit/chainlink/v2/core/services/pg"
	"github.com/smartcontractkit/chainlink/v2/core/utils"
)

// // TODO: needs to be populated asynchronously from onchain ConfigurationStore
// type ChannelDefinitionCache interface {
//     // TODO: Would this necessarily need to be scoped by contract address?
//     Definitions() ChannelDefinitions
//     services.Service
// }

type ChannelDefinitionCacheORM interface {
	// TODO: What about delete/cleanup?
	LoadChannelDefinitions(ctx context.Context, addr common.Address) (cd commontypes.ChannelDefinitions, blockNum int64, err error)
	StoreChannelDefinitions(ctx context.Context, cd commontypes.ChannelDefinitions, blockNum int64) (err error)
}

var channelConfigStoreABI abi.ABI

func init() {
	var err error
	channelConfigStoreABI, err = abi.JSON(strings.NewReader(channel_config_store.ChannelConfigStoreABI))
	if err != nil {
		panic(err)
	}
}

var _ commontypes.ChannelDefinitionCache = &channelDefinitionCache{}

type channelDefinitionCache struct {
	services.StateMachine

	orm ChannelDefinitionCacheORM

	filterName string
	lp         logpoller.LogPoller
	fromBlock  int64
	addr       common.Address
	lggr       logger.Logger

	definitionsMu       sync.RWMutex
	definitions         commontypes.ChannelDefinitions
	definitionsBlockNum int64

	wg     sync.WaitGroup
	chStop chan struct{}
}

var (
	topicNewChannelDefinition     = (channel_config_store.ChannelConfigStoreNewChannelDefinition{}).Topic()
	topicChannelDefinitionRemoved = (channel_config_store.ChannelConfigStoreChannelDefinitionRemoved{}).Topic()

	allTopics = []common.Hash{topicNewChannelDefinition, topicChannelDefinitionRemoved}
)

func NewChannelDefinitionCache(lggr logger.Logger, orm ChannelDefinitionCacheORM, lp logpoller.LogPoller, addr common.Address, fromBlock int64) commontypes.ChannelDefinitionCache {
	filterName := logpoller.FilterName("OCR3 LLO ChannelDefinitionCachePoller", addr.String())
	return &channelDefinitionCache{
		services.StateMachine{},
		orm,
		filterName,
		lp,
		0, // TODO: fromblock needs to be loaded from DB cache somehow because we don't want to scan all logs every time we start this job
		addr,
		// TODO: Does it log chain ID?
		lggr.Named("ChannelDefinitionCache").With("addr", addr, "fromBlock", fromBlock),
		sync.RWMutex{},
		nil,
		fromBlock,
		sync.WaitGroup{},
		make(chan struct{}),
	}
}

// TODO: Needs a way to subscribe/unsubscribe to contracts

func (c *channelDefinitionCache) Start(ctx context.Context) error {
	// Initial load from DB, then async poll from chain thereafter
	// TODO: needs to be populated asynchronously from onchain ConfigurationStore
	return c.StartOnce("ChannelDefinitionCache", func() (err error) {
		err = c.lp.RegisterFilter(logpoller.Filter{Name: c.filterName, EventSigs: allTopics, Addresses: []common.Address{c.addr}}, pg.WithParentCtx(ctx))
		if err != nil {
			return err
		}
		if definitions, definitionsBlockNum, err := c.orm.LoadChannelDefinitions(ctx, c.addr); err != nil {
			return err
		} else if definitions != nil {
			c.definitions = definitions
			c.definitionsBlockNum = definitionsBlockNum
		} else {
			// ensure non-nil map ready for assignment later
			c.definitions = make(commontypes.ChannelDefinitions)
			// leave c.definitionsBlockNum as provided fromBlock argument
		}
		c.wg.Add(1)
		go c.poll()
		return nil
	})
}

// TODO: make this configurable?
const pollInterval = 1 * time.Second

func (c *channelDefinitionCache) poll() {
	defer c.wg.Done()

	pollT := time.NewTicker(utils.WithJitter(pollInterval))

	for {
		select {
		case <-c.chStop:
			return
		case <-pollT.C:
			if n, err := c.fetchFromChain(); err != nil {
				// TODO: retry with backoff?
				panic(err)
			} else {
				if n > 0 {
					c.lggr.Infow("Updated channel definitions", "nLogs", n, "definitionsBlockNum", c.definitionsBlockNum)
				} else {
					c.lggr.Debugw("No new channel definitions", "nLogs", 0, "definitionsBlockNum", c.definitionsBlockNum)
				}
			}
		}
	}
}

func (c *channelDefinitionCache) fetchFromChain() (nLogs int, err error) {
	// TODO: Pass context
	latest, err := c.lp.LatestBlock()
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	} else if err != nil {
		return 0, err
	}
	toBlock := latest.BlockNumber

	fromBlock := c.definitionsBlockNum

	if toBlock <= fromBlock {
		return 0, nil
	}

	ctx, cancel := services.StopChan(c.chStop).NewCtx()
	defer cancel()
	// NOTE: We assume that log poller returns logs in ascending order chronologically
	logs, err := c.lp.LogsWithSigs(fromBlock, toBlock, allTopics, c.addr, pg.WithParentCtx(ctx))
	if err != nil {
		// TODO: retry?
		return 0, err
	}
	for _, log := range logs {
		if err := c.applyLog(log); err != nil {
			return 0, err
		}
	}

	// Use context.Background() here because we want to try to save even if we
	// are closing
	if err = c.orm.StoreChannelDefinitions(context.Background(), c.Definitions(), toBlock); err != nil {
		return 0, err
	}

	c.definitionsBlockNum = toBlock

	return len(logs), nil
}

func (c *channelDefinitionCache) applyLog(log logpoller.Log) error {
	switch log.EventSig {
	case topicNewChannelDefinition:
		unpacked := new(channel_config_store.ChannelConfigStoreNewChannelDefinition)

		err := channelConfigStoreABI.UnpackIntoInterface(unpacked, "NewChannelDefinition", log.Data)
		if err != nil {
			return fmt.Errorf("failed to unpack log data: %w", err)
		}

		c.applyNewChannelDefinition(unpacked)
	case topicChannelDefinitionRemoved:
		unpacked := new(channel_config_store.ChannelConfigStoreChannelDefinitionRemoved)

		err := channelConfigStoreABI.UnpackIntoInterface(unpacked, "ChannelDefinitionRemoved", log.Data)
		if err != nil {
			return fmt.Errorf("failed to unpack log data: %w", err)
		}

		c.applyChannelDefinitionRemoved(unpacked)
	default:
		panic("TODO")
	}
	return nil
}

func (c *channelDefinitionCache) applyNewChannelDefinition(log *channel_config_store.ChannelConfigStoreNewChannelDefinition) {
	rf := DecodeReportFormat(log.ChannelDefinition.ReportFormat)
	streamIDs := make([]commontypes.StreamID, len(log.ChannelDefinition.StreamIDs))
	for i, streamID := range log.ChannelDefinition.StreamIDs {
		streamIDs[i] = streamID
	}
	c.definitionsMu.Lock()
	defer c.definitionsMu.Unlock()
	c.definitions[log.ChannelId] = commontypes.ChannelDefinition{
		ReportFormat:  rf,
		ChainSelector: log.ChannelDefinition.ChainSelector,
		StreamIDs:     streamIDs,
	}
}

func DecodeReportFormat(onchainRF [8]byte) commontypes.LLOReportFormat {
	n := bytes.Index(onchainRF[:], []byte{0})
	if n < 0 {
		n = len(onchainRF)
	}
	rf := string(onchainRF[:n])
	return commontypes.LLOReportFormat(rf)
}

func (c *channelDefinitionCache) applyChannelDefinitionRemoved(log *channel_config_store.ChannelConfigStoreChannelDefinitionRemoved) {
	c.definitionsMu.Lock()
	defer c.definitionsMu.Unlock()
	delete(c.definitions, log.ChannelId)
}

func (c *channelDefinitionCache) Close() error {
	// TODO
	// TODO: unregister filter (on job delete)?
	return c.StopOnce("ChannelDefinitionCache", func() error {
		close(c.chStop)
		c.wg.Wait()
		return nil
	})
}

func (c *channelDefinitionCache) HealthReport() map[string]error {
	report := map[string]error{c.Name(): c.Healthy()}
	return report
}

func (c *channelDefinitionCache) Name() string { return c.lggr.Name() }

func (c *channelDefinitionCache) Definitions() commontypes.ChannelDefinitions {
	c.definitionsMu.RLock()
	defer c.definitionsMu.RUnlock()
	return maps.Clone(c.definitions)
}
