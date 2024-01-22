package gas_test

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common/hexutil"
	pkgerrors "github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/services/servicetest"
	"github.com/smartcontractkit/chainlink/v2/core/chains/evm/assets"
	"github.com/smartcontractkit/chainlink/v2/core/chains/evm/gas"
	"github.com/smartcontractkit/chainlink/v2/core/chains/evm/gas/mocks"
	"github.com/smartcontractkit/chainlink/v2/core/internal/testutils"
)

func TestSuggestedPriceEstimator(t *testing.T) {
	t.Parallel()

	maxGasPrice := assets.NewWeiI(100)

	calldata := []byte{0x00, 0x00, 0x01, 0x02, 0x03}
	const gasLimit uint32 = 80000

	t.Run("calling GetLegacyGas on unstarted estimator returns error", func(t *testing.T) {
		client := mocks.NewRPCClient(t)
		o := gas.NewSuggestedPriceEstimator(logger.Test(t), client)
		_, _, err := o.GetLegacyGas(testutils.Context(t), calldata, gasLimit, maxGasPrice)
		assert.EqualError(t, err, "estimator is not started")
	})

	t.Run("calling GetLegacyGas on started estimator returns prices", func(t *testing.T) {
		client := mocks.NewRPCClient(t)
		client.On("CallContext", mock.Anything, mock.Anything, "eth_gasPrice").Return(nil).Run(func(args mock.Arguments) {
			res := args.Get(1).(*hexutil.Big)
			(*big.Int)(res).SetInt64(42)
		})

		o := gas.NewSuggestedPriceEstimator(logger.Test(t), client)
		servicetest.RunHealthy(t, o)
		gasPrice, chainSpecificGasLimit, err := o.GetLegacyGas(testutils.Context(t), calldata, gasLimit, maxGasPrice)
		require.NoError(t, err)
		assert.Equal(t, assets.NewWeiI(42), gasPrice)
		assert.Equal(t, gasLimit, chainSpecificGasLimit)
	})

	t.Run("gas price is lower than user specified max gas price", func(t *testing.T) {
		client := mocks.NewRPCClient(t)
		o := gas.NewSuggestedPriceEstimator(logger.Test(t), client)

		client.On("CallContext", mock.Anything, mock.Anything, "eth_gasPrice").Return(nil).Run(func(args mock.Arguments) {
			res := args.Get(1).(*hexutil.Big)
			(*big.Int)(res).SetInt64(42)
		})

		servicetest.RunHealthy(t, o)
		gasPrice, chainSpecificGasLimit, err := o.GetLegacyGas(testutils.Context(t), calldata, gasLimit, assets.NewWeiI(40))
		require.Error(t, err)
		assert.EqualError(t, err, "estimated gas price: 42 wei is greater than the maximum gas price configured: 40 wei")
		assert.Nil(t, gasPrice)
		assert.Equal(t, uint32(0), chainSpecificGasLimit)
	})

	t.Run("gas price is lower than global max gas price", func(t *testing.T) {
		client := mocks.NewRPCClient(t)
		o := gas.NewSuggestedPriceEstimator(logger.Test(t), client)

		client.On("CallContext", mock.Anything, mock.Anything, "eth_gasPrice").Return(nil).Run(func(args mock.Arguments) {
			res := args.Get(1).(*hexutil.Big)
			(*big.Int)(res).SetInt64(120)
		})

		servicetest.RunHealthy(t, o)
		gasPrice, chainSpecificGasLimit, err := o.GetLegacyGas(testutils.Context(t), calldata, gasLimit, assets.NewWeiI(110))
		assert.EqualError(t, err, "estimated gas price: 120 wei is greater than the maximum gas price configured: 110 wei")
		assert.Nil(t, gasPrice)
		assert.Equal(t, uint32(0), chainSpecificGasLimit)
	})

	t.Run("calling BumpLegacyGas always returns error", func(t *testing.T) {
		client := mocks.NewRPCClient(t)
		o := gas.NewSuggestedPriceEstimator(logger.Test(t), client)
		_, _, err := o.BumpLegacyGas(testutils.Context(t), assets.NewWeiI(42), gasLimit, assets.NewWeiI(10), nil)
		assert.EqualError(t, err, "bump gas is not supported for this chain")
	})

	t.Run("calling GetLegacyGas on started estimator if initial call failed returns error", func(t *testing.T) {
		client := mocks.NewRPCClient(t)
		o := gas.NewSuggestedPriceEstimator(logger.Test(t), client)

		client.On("CallContext", mock.Anything, mock.Anything, "eth_gasPrice").Return(pkgerrors.New("kaboom"))

		servicetest.RunHealthy(t, o)

		_, _, err := o.GetLegacyGas(testutils.Context(t), calldata, gasLimit, maxGasPrice)
		assert.EqualError(t, err, "failed to estimate gas; gas price not set")
	})
}
