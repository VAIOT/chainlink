package evm

import (
	"encoding/json"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"github.com/smartcontractkit/libocr/gethwrappers2/ocr2aggregator"
	"github.com/smartcontractkit/libocr/offchainreporting2plus/chains/evmutil"
	"github.com/smartcontractkit/libocr/offchainreporting2plus/types"

	relaytypes "github.com/smartcontractkit/chainlink-relay/pkg/types"
	"github.com/smartcontractkit/chainlink/v2/core/chains/evm"
	"github.com/smartcontractkit/chainlink/v2/core/logger"
	"github.com/smartcontractkit/chainlink/v2/core/services/keystore"
	functionsConfig "github.com/smartcontractkit/chainlink/v2/core/services/ocr2/plugins/functions/config"
	functionsRelay "github.com/smartcontractkit/chainlink/v2/core/services/relay/evm/functions"
	evmRelayTypes "github.com/smartcontractkit/chainlink/v2/core/services/relay/evm/types"
)

type functionsProvider struct {
	*configWatcher
	contractTransmitter ContractTransmitter
}

var (
	_ relaytypes.Plugin = (*functionsProvider)(nil)
)

func (p *functionsProvider) ContractTransmitter() types.ContractTransmitter {
	return p.contractTransmitter
}

func NewFunctionsProvider(chainSet evm.ChainSet, rargs relaytypes.RelayArgs, pargs relaytypes.PluginArgs, lggr logger.Logger, ethKeystore keystore.Eth, pluginType functionsRelay.FunctionsPluginType) (relaytypes.Plugin, error) {
	configWatcher, err := newFunctionsConfigProvider(pluginType, chainSet, rargs, pargs, lggr)
	if err != nil {
		return nil, err
	}

	// TODO: use custom transmitter here
	contractTransmitter, err := newContractTransmitter(lggr, rargs, pargs.TransmitterID, configWatcher, ethKeystore)
	if err != nil {
		return nil, err
	}
	return &functionsProvider{
		configWatcher:       configWatcher,
		contractTransmitter: contractTransmitter,
	}, nil
}

func newFunctionsConfigProvider(pluginType functionsRelay.FunctionsPluginType, chainSet evm.ChainSet, rargs relaytypes.RelayArgs, pargs relaytypes.PluginArgs, lggr logger.Logger) (*configWatcher, error) {
	var pluginConfig functionsConfig.PluginConfig
	err := json.Unmarshal(pargs.PluginConfig, &pluginConfig)
	if err != nil {
		return nil, err
	}
	var relayConfig evmRelayTypes.RelayConfig
	err = json.Unmarshal(rargs.RelayConfig, &relayConfig)
	if err != nil {
		return nil, err
	}
	chain, err := chainSet.Get(relayConfig.ChainID.ToInt())
	if err != nil {
		return nil, err
	}
	if !common.IsHexAddress(rargs.ContractID) {
		return nil, errors.Errorf("invalid contractID, expected hex address")
	}

	contractAddress := common.HexToAddress(rargs.ContractID)
	contractABI, err := abi.JSON(strings.NewReader(ocr2aggregator.OCR2AggregatorMetaData.ABI))
	if err != nil {
		return nil, errors.Wrap(err, "could not get contract ABI JSON")
	}

	cp, err := functionsRelay.NewFunctionsConfigPoller(pluginType, chain.LogPoller(), contractAddress, pluginConfig.ContractVersion, lggr)
	if err != nil {
		return nil, err
	}

	offchainConfigDigester := functionsRelay.FunctionsOffchainConfigDigester{
		PluginType: pluginType,
		BaseDigester: evmutil.EVMOffchainConfigDigester{
			ChainID:         chain.ID().Uint64(),
			ContractAddress: contractAddress,
		},
	}

	return newConfigWatcher(lggr, contractAddress, contractABI, offchainConfigDigester, cp, chain, relayConfig.FromBlock, rargs.New), nil
}
