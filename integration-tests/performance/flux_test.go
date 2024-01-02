package performance

import (
	"fmt"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-testing-framework/blockchain"
	ctfClient "github.com/smartcontractkit/chainlink-testing-framework/client"
	ctf_config "github.com/smartcontractkit/chainlink-testing-framework/config"
	"github.com/smartcontractkit/chainlink-testing-framework/k8s/environment"
	"github.com/smartcontractkit/chainlink-testing-framework/k8s/pkg/helm/chainlink"
	"github.com/smartcontractkit/chainlink-testing-framework/k8s/pkg/helm/ethereum"
	"github.com/smartcontractkit/chainlink-testing-framework/k8s/pkg/helm/mockserver"
	mockservercfg "github.com/smartcontractkit/chainlink-testing-framework/k8s/pkg/helm/mockserver-cfg"
	"github.com/smartcontractkit/chainlink-testing-framework/logging"
	"github.com/smartcontractkit/chainlink-testing-framework/networks"
	"github.com/smartcontractkit/chainlink-testing-framework/utils/testcontext"

	"github.com/smartcontractkit/chainlink/integration-tests/actions"
	"github.com/smartcontractkit/chainlink/integration-tests/client"
	"github.com/smartcontractkit/chainlink/integration-tests/contracts"
	tc "github.com/smartcontractkit/chainlink/integration-tests/testconfig"
	"github.com/smartcontractkit/chainlink/integration-tests/testsetups"
)

func TestFluxPerformance(t *testing.T) {
	l := logging.GetTestLogger(t)
	config, err := tc.GetConfig("Performance", tc.DirectRequest)
	if err != nil {
		t.Fatal(err)
	}

	testEnvironment, testNetwork := setupFluxTest(t, &config)
	if testEnvironment.WillUseRemoteRunner() {
		return
	}

	chainClient, err := blockchain.NewEVMClient(testNetwork, testEnvironment, l)
	require.NoError(t, err, "Connecting to blockchain nodes shouldn't fail")
	contractDeployer, err := contracts.NewContractDeployer(chainClient, l)
	require.NoError(t, err, "Deploying contracts shouldn't fail")
	chainlinkNodes, err := client.ConnectChainlinkNodes(testEnvironment)
	require.NoError(t, err, "Connecting to chainlink nodes shouldn't fail")
	nodeAddresses, err := actions.ChainlinkNodeAddresses(chainlinkNodes)
	require.NoError(t, err, "Retreiving on-chain wallet addresses for chainlink nodes shouldn't fail")
	mockServer, err := ctfClient.ConnectMockServer(testEnvironment)
	require.NoError(t, err, "Creating mock server client shouldn't fail")

	chainClient.ParallelTransactions(true)

	adapterUUID := uuid.New().String()
	adapterPath := fmt.Sprintf("/variable-%s", adapterUUID)
	err = mockServer.SetValuePath(adapterPath, 1e5)
	require.NoError(t, err, "Setting mockserver value path shouldn't fail")

	linkToken, err := contractDeployer.DeployLinkTokenContract()
	require.NoError(t, err, "Deploying Link Token Contract shouldn't fail")
	fluxInstance, err := contractDeployer.DeployFluxAggregatorContract(linkToken.Address(), contracts.DefaultFluxAggregatorOptions())
	require.NoError(t, err, "Deploying Flux Aggregator Contract shouldn't fail")
	err = chainClient.WaitForEvents()
	require.NoError(t, err, "Failed waiting for deployment of flux aggregator contract")

	err = linkToken.Transfer(fluxInstance.Address(), big.NewInt(1e18))
	require.NoError(t, err, "Funding Flux Aggregator Contract shouldn't fail")
	err = chainClient.WaitForEvents()
	require.NoError(t, err, "Failed waiting for funding of flux aggregator contract")

	err = fluxInstance.UpdateAvailableFunds()
	require.NoError(t, err, "Updating the available funds on the Flux Aggregator Contract shouldn't fail")

	err = actions.FundChainlinkNodes(chainlinkNodes, chainClient, big.NewFloat(.02))
	require.NoError(t, err, "Funding chainlink nodes with ETH shouldn't fail")

	err = fluxInstance.SetOracles(
		contracts.FluxAggregatorSetOraclesOptions{
			AddList:            nodeAddresses,
			RemoveList:         []common.Address{},
			AdminList:          nodeAddresses,
			MinSubmissions:     3,
			MaxSubmissions:     3,
			RestartDelayRounds: 0,
		})
	require.NoError(t, err, "Setting oracle options in the Flux Aggregator contract shouldn't fail")
	err = chainClient.WaitForEvents()
	require.NoError(t, err, "Waiting for event subscriptions in nodes shouldn't fail")
	oracles, err := fluxInstance.GetOracles(testcontext.Get(t))
	require.NoError(t, err, "Getting oracle details from the Flux aggregator contract shouldn't fail")
	l.Info().Str("Oracles", strings.Join(oracles, ",")).Msg("Oracles set")

	adapterFullURL := fmt.Sprintf("%s%s", mockServer.Config.ClusterURL, adapterPath)
	bta := &client.BridgeTypeAttributes{
		Name: fmt.Sprintf("variable-%s", adapterUUID),
		URL:  adapterFullURL,
	}
	for i, n := range chainlinkNodes {
		err = n.MustCreateBridge(bta)
		require.NoError(t, err, "Creating bridge shouldn't fail for node %d", i+1)

		fluxSpec := &client.FluxMonitorJobSpec{
			Name:              fmt.Sprintf("flux-monitor-%s", adapterUUID),
			ContractAddress:   fluxInstance.Address(),
			EVMChainID:        chainClient.GetChainID().String(),
			Threshold:         0,
			AbsoluteThreshold: 0,
			PollTimerPeriod:   15 * time.Second, // min 15s
			IdleTimerDisabled: true,
			ObservationSource: client.ObservationSourceSpecBridge(bta),
		}
		_, err = n.MustCreateJob(fluxSpec)
		require.NoError(t, err, "Creating flux job shouldn't fail for node %d", i+1)
	}

	profileFunction := func(chainlinkNode *client.ChainlinkClient) {
		if chainlinkNode != chainlinkNodes[len(chainlinkNodes)-1].ChainlinkClient {
			// Not the last node, hence not all nodes started profiling yet.
			return
		}
		fluxRoundTimeout := 2 * time.Minute
		fluxRound := contracts.NewFluxAggregatorRoundConfirmer(fluxInstance, big.NewInt(1), fluxRoundTimeout, l)
		chainClient.AddHeaderEventSubscription(fluxInstance.Address(), fluxRound)
		err = chainClient.WaitForEvents()
		require.NoError(t, err, "Waiting for event subscriptions in nodes shouldn't fail")
		data, err := fluxInstance.GetContractData(testcontext.Get(t))
		require.NoError(t, err, "Getting contract data from flux aggregator contract shouldn't fail")
		l.Info().Interface("Data", data).Msg("Round data")
		require.Equal(t, int64(1e5), data.LatestRoundData.Answer.Int64(),
			"Expected latest round answer to be %d, but found %d", int64(1e5), data.LatestRoundData.Answer.Int64())
		require.Equal(t, int64(1), data.LatestRoundData.RoundId.Int64(),
			"Expected latest round id to be %d, but found %d", int64(1), data.LatestRoundData.RoundId.Int64())
		require.Equal(t, int64(1), data.LatestRoundData.AnsweredInRound.Int64(),
			"Expected latest round's answered in round to be %d, but found %d", int64(1), data.LatestRoundData.AnsweredInRound.Int64())
		require.Equal(t, int64(999999999999999997), data.AvailableFunds.Int64(),
			"Expected available funds to be %d, but found %d", int64(999999999999999997), data.AvailableFunds.Int64())
		require.Equal(t, int64(3), data.AllocatedFunds.Int64(),
			"Expected allocated funds to be %d, but found %d", int64(3), data.AllocatedFunds.Int64())

		fluxRound = contracts.NewFluxAggregatorRoundConfirmer(fluxInstance, big.NewInt(2), fluxRoundTimeout, l)
		chainClient.AddHeaderEventSubscription(fluxInstance.Address(), fluxRound)
		err = mockServer.SetValuePath(adapterPath, 1e10)
		require.NoError(t, err, "Setting value path in mock server shouldn't fail")
		err = chainClient.WaitForEvents()
		require.NoError(t, err, "Waiting for event subscriptions in nodes shouldn't fail")
		data, err = fluxInstance.GetContractData(testcontext.Get(t))
		require.NoError(t, err, "Getting contract data from flux aggregator contract shouldn't fail")
		require.Equal(t, int64(1e10), data.LatestRoundData.Answer.Int64(),
			"Expected latest round answer to be %d, but found %d", int64(1e10), data.LatestRoundData.Answer.Int64())
		require.Equal(t, int64(2), data.LatestRoundData.RoundId.Int64(),
			"Expected latest round id to be %d, but found %d", int64(2), data.LatestRoundData.RoundId.Int64())
		require.Equal(t, int64(999999999999999994), data.AvailableFunds.Int64(),
			"Expected available funds to be %d, but found %d", int64(999999999999999994), data.AvailableFunds.Int64())
		require.Equal(t, int64(6), data.AllocatedFunds.Int64(),
			"Expected allocated funds to be %d, but found %d", int64(6), data.AllocatedFunds.Int64())
		l.Info().Interface("data", data).Msg("Round data")

		for _, oracleAddr := range nodeAddresses {
			payment, _ := fluxInstance.WithdrawablePayment(testcontext.Get(t), oracleAddr)
			require.Equal(t, int64(2), payment.Int64(),
				"Expected flux aggregator contract's withdrawable payment to be %d, but found %d", int64(2), payment.Int64())
		}
	}

	profileTest := testsetups.NewChainlinkProfileTest(testsetups.ChainlinkProfileTestInputs{
		ProfileFunction: profileFunction,
		ProfileDuration: 30 * time.Second,
		ChainlinkNodes:  chainlinkNodes,
	})
	// Register cleanup
	t.Cleanup(func() {
		CleanupPerformanceTest(t, testEnvironment, chainlinkNodes, profileTest.TestReporter, &config, chainClient)
	})
	profileTest.Setup(testEnvironment)
	profileTest.Run()
}

func setupFluxTest(t *testing.T, config *tc.TestConfig) (testEnvironment *environment.Environment, testNetwork blockchain.EVMNetwork) {
	testNetwork = networks.MustGetSelectedNetworkConfig(config.Network)[0]
	evmConf := ethereum.New(nil)
	if !testNetwork.Simulated {
		evmConf = ethereum.New(&ethereum.Props{
			NetworkName: testNetwork.Name,
			Simulated:   testNetwork.Simulated,
			WsURLs:      testNetwork.URLs,
		})
	}
	baseTOML := `[WebServer]
HTTPWriteTimout = '300s'

[OCR]
Enabled = true`

	var overrideFn = func(_ interface{}, target interface{}) {
		ctf_config.MustConfigOverrideChainlinkVersion(config.ChainlinkImage, target)
	}

	cd := chainlink.NewWithOverride(0, map[string]interface{}{
		"replicas": 3,
		"toml":     networks.AddNetworksConfig(baseTOML, config.Pyroscope, testNetwork),
	}, config.ChainlinkImage, overrideFn)

	testEnvironment = environment.New(&environment.Config{
		NamespacePrefix:    fmt.Sprintf("performance-flux-%s", strings.ReplaceAll(strings.ToLower(testNetwork.Name), " ", "-")),
		Test:               t,
		PreventPodEviction: true,
	}).
		AddHelm(mockservercfg.New(nil)).
		AddHelm(mockserver.New(nil)).
		AddHelm(evmConf).
		AddHelm(cd)
	err := testEnvironment.Run()
	require.NoError(t, err, "Error running test environment")
	return testEnvironment, testNetwork
}
