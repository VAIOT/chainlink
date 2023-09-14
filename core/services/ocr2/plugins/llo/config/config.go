// config is a separate package so that we can validate
// the config in other packages, for example in job at job create time.

package config

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"

	"github.com/ethereum/go-ethereum/common"
	pkgerrors "github.com/pkg/errors"

	commontypes "github.com/smartcontractkit/chainlink-common/pkg/types"

	"github.com/smartcontractkit/chainlink/v2/core/utils"
)

type PluginConfig struct {
	RawServerURL string              `json:"serverURL" toml:"serverURL"`
	ServerPubKey utils.PlainHexBytes `json:"serverPubKey" toml:"serverPubKey"`

	ChannelDefinitionsContractAddress   common.Address `json:"channelDefinitionsContractAddress" toml:"channelDefinitionsContractAddress"`
	ChannelDefinitionsContractFromBlock int64          `json:"channelDefinitionsContractFromBlock" toml:"channelDefinitionsContractFromBlock"`

	// NOTE: ChannelDefinitions is an override.
	// If ChannelDefinitions is specified, values for
	// ChannelDefinitionsContractAddress and
	// ChannelDefinitionsContractFromBlock will be ignored
	ChannelDefinitions commontypes.ChannelDefinitions `json:"channelDefinitions" toml:"channelDefinitions"`
}

func (p PluginConfig) Validate() (merr error) {
	if p.RawServerURL == "" {
		merr = errors.New("llo: ServerURL must be specified")
	} else {
		var normalizedURI string
		if schemeRegexp.MatchString(p.RawServerURL) {
			normalizedURI = p.RawServerURL
		} else {
			normalizedURI = fmt.Sprintf("wss://%s", p.RawServerURL)
		}
		uri, err := url.ParseRequestURI(normalizedURI)
		if err != nil {
			merr = pkgerrors.Wrap(err, "llo: invalid value for ServerURL")
		} else if uri.Scheme != "wss" {
			merr = pkgerrors.Errorf(`llo: invalid scheme specified for MercuryServer, got: %q (scheme: %q) but expected a websocket url e.g. "192.0.2.2:4242" or "wss://192.0.2.2:4242"`, p.RawServerURL, uri.Scheme)
		}
	}

	if p.ChannelDefinitions != nil {
		if p.ChannelDefinitionsContractAddress != (common.Address{}) {
			merr = errors.Join(merr, errors.New("llo: ChannelDefinitionsContractAddress is not allowed if ChannelDefinitions is specified"))
		}
		if p.ChannelDefinitionsContractFromBlock != 0 {
			merr = errors.Join(merr, errors.New("llo: ChannelDefinitionsContractFromBlock is not allowed if ChannelDefinitions is specified"))
		}
	} else {
		if p.ChannelDefinitionsContractAddress == (common.Address{}) {
			merr = errors.Join(merr, errors.New("llo: ChannelDefinitionsContractAddress is required if ChannelDefinitions is not specified"))
		}
		if p.ChannelDefinitionsContractFromBlock == 0 {
			merr = errors.Join(merr, errors.New("llo: ChannelDefinitionsContractFromBlock is required if ChannelDefinitions is not specified"))
		}
	}

	if len(p.ServerPubKey) != 32 {
		merr = errors.Join(merr, errors.New("llo: ServerPubKey is required and must be a 32-byte hex string"))
	}

	return merr
}

var schemeRegexp = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9+.-]*://`)
var wssRegexp = regexp.MustCompile(`^wss://`)

func (p PluginConfig) ServerURL() string {
	return wssRegexp.ReplaceAllString(p.RawServerURL, "")
}
