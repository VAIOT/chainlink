package llo

import (
	"testing"

	commontypes "github.com/smartcontractkit/chainlink-common/pkg/types"
	"github.com/stretchr/testify/assert"
)

func Test_ChannelDefinitionCache(t *testing.T) {
	t.Run("Definitions", func(t *testing.T) {
		t.Fatal("TODO")
	})
}

func Test_ChannelDefinitionCache_DecodeReportFormat(t *testing.T) {
	type tc struct {
		name     string
		rf       [8]byte
		expected string
	}
	tcs := []tc{
		{
			"normal",
			[8]byte{'e', 'v', 'm'},
			"evm",
		},
		{
			"empty",
			[8]byte{},
			"",
		},
		{
			"max length",
			[8]byte{'e', 'v', 'm', '1', '2', '3', '4', '5'},
			"evm12345",
		},
		{
			"unprintable characters",
			[8]byte{1, 2, 3, 4},
			"\x01\x02\x03\x04",
		},
		{
			"max length unprintable characters",
			[8]byte{1, 2, 3, 4, 5, 6, 7, 8},
			"\x01\x02\x03\x04\x05\x06\a\b",
		},
	}
	for _, testcase := range tcs {
		t.Run(testcase.name, func(t *testing.T) {
			assert.Equal(t, commontypes.LLOReportFormat(testcase.expected), DecodeReportFormat(testcase.rf))
		})
	}
}
