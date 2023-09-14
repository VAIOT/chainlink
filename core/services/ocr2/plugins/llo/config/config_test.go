package config

import (
	"testing"

	"github.com/pelletier/go-toml/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Config(t *testing.T) {
	t.Run("unmarshals from toml", func(t *testing.T) {
		t.Run("with all possible values set", func(t *testing.T) {
			rawToml := `
				ServerURL = "example.com:80"
				ServerPubKey = "724ff6eae9e900270edfff233e16322a70ec06e1a6e62a81ef13921f398f6c93"
				ChannelDefinitionsContractAddress = "0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
				ChannelDefinitionsContractFromBlock = 1234
				ChannelDefinitions = """
{
	"42": {
		"reportFormat": "example-llo-report-format",
		"chainSelector": 142,
		"streamIds": [1, 2]
	},
	"43": {
		"reportFormat": "example-llo-report-format",
		"chainSelector": 142,
		"streamIds": [1, 3]
	}
	"44": {
		"reportFormat": "example-llo-report-format",
		"chainSelector": 143,
		"streamIds": [1, 4]
	}
}
"""
			`

			var mc PluginConfig
			err := toml.Unmarshal([]byte(rawToml), &mc)
			require.NoError(t, err)

			assert.Equal(t, "example.com:80", mc.RawServerURL)
			assert.Equal(t, "724ff6eae9e900270edfff233e16322a70ec06e1a6e62a81ef13921f398f6c93", mc.ServerPubKey.String())
			assert.Equal(t, "foo", mc.ChannelDefinitionsContractAddress)
			assert.Equal(t, int64(1234), mc.ChannelDefinitionsContractFromBlock)
			assert.Equal(t, "foo", mc.ChannelDefinitions)

			err = mc.Validate()
			require.NoError(t, err)
		})

		t.Run("with invalid values", func(t *testing.T) {
			rawToml := `
				InitialBlockNumber = "invalid"
			`

			var mc PluginConfig
			err := toml.Unmarshal([]byte(rawToml), &mc)
			require.Error(t, err)
			assert.EqualError(t, err, `toml: strconv.ParseInt: parsing "invalid": invalid syntax`)

			rawToml = `
				ServerURL = "http://example.com"
				ServerPubKey = "4242"
			`

			err = toml.Unmarshal([]byte(rawToml), &mc)
			require.NoError(t, err)

			err = mc.Validate()
			require.Error(t, err)
			assert.Contains(t, err.Error(), `Mercury: invalid scheme specified for MercuryServer, got: "http://example.com" (scheme: "http") but expected a websocket url e.g. "192.0.2.2:4242" or "wss://192.0.2.2:4242"`)
			assert.Contains(t, err.Error(), `mercury: ServerPubKey is required and must be a 32-byte hex string`)
		})
	})
}
