package automation

import (
	"errors"
	"math/big"
)

type Config struct {
	General *General `toml:"General"`
	Load    []Load   `toml:"Load"`
}

func (c *Config) ApplyOverrides(from *Config) error {
	if from == nil {
		return nil
	}
	if from.General != nil && c.General == nil {
		c.General = from.General
	} else if from.General != nil && c.General != nil {
		if err := c.General.ApplyOverrides(from.General); err != nil {
			return err
		}
	}
	if len(from.Load) > 0 && len(c.Load) == 0 {
		c.Load = from.Load
	} else if len(from.Load) > 0 && len(c.Load) > 0 {
		from.Load = append(from.Load, c.Load...)
	}

	return nil
}

func (c *Config) Validate() error {
	if c.General != nil {
		if err := c.General.Validate(); err != nil {
			return err
		}
	}
	if len(c.Load) > 0 {
		for _, load := range c.Load {
			if err := load.Validate(); err != nil {
				return err
			}
		}
	}
	return nil
}

// General is a common configuration for all automation performance tests
type General struct {
	NumberOfNodes         *int    `toml:"number_of_nodes"`
	Duration              *int    `toml:"duration"`
	BlockTime             *int    `toml:"block_time"`
	SpecType              *string `toml:"spec_type"`
	ChainlinkNodeLogLevel *string `toml:"chainlink_node_log_level"`
	UsePrometheus         *bool   `toml:"use_prometheus"`
}

func (c *General) ApplyOverrides(from *General) error {
	if from == nil {
		return nil
	}
	if from.NumberOfNodes != nil {
		c.NumberOfNodes = from.NumberOfNodes
	}
	if from.Duration != nil {
		c.Duration = from.Duration
	}
	if from.BlockTime != nil {
		c.BlockTime = from.BlockTime
	}
	if from.SpecType != nil {
		c.SpecType = from.SpecType
	}
	if from.ChainlinkNodeLogLevel != nil {
		c.ChainlinkNodeLogLevel = from.ChainlinkNodeLogLevel
	}
	if from.UsePrometheus != nil {
		c.UsePrometheus = from.UsePrometheus
	}

	return nil
}

func (c *General) Validate() error {
	if c.NumberOfNodes == nil || *c.NumberOfNodes < 1 {
		return errors.New("number_of_nodes must be set to a positive integer")
	}
	if c.Duration == nil || *c.Duration < 1 {
		return errors.New("duration must be set to a positive integer")
	}
	if c.BlockTime == nil || *c.BlockTime < 1 {
		return errors.New("block_time must be set to a positive integer")
	}
	if c.SpecType == nil {
		return errors.New("spec_type must be set")
	}
	if c.ChainlinkNodeLogLevel == nil {
		return errors.New("chainlink_node_log_level must be set")
	}

	return nil
}

type Load struct {
	NumberOfUpkeeps               *int     `toml:"number_of_upkeeps"`
	NumberOfEvents                *int     `toml:"number_of_events"`
	NumberOfSpamMatchingEvents    *int     `toml:"number_of_spam_matching_events"`
	NumberOfSpamNonMatchingEvents *int     `toml:"number_of_spam_non_matching_events"`
	CheckBurnAmount               *big.Int `toml:"check_burn_amount"`
	PerformBurnAmount             *big.Int `toml:"perform_burn_amount"`
	SharedTrigger                 *bool    `toml:"shared_trigger"`
	UpkeepGasLimit                *uint32  `toml:"upkeep_gas_limit"`
}

func (c *Load) ApplyOverrides(from *Load) error {
	if from == nil {
		return nil
	}
	if from.NumberOfUpkeeps != nil {
		c.NumberOfUpkeeps = from.NumberOfUpkeeps
	}
	if from.NumberOfEvents != nil {
		c.NumberOfEvents = from.NumberOfEvents
	}
	if from.NumberOfSpamMatchingEvents != nil {
		c.NumberOfSpamMatchingEvents = from.NumberOfSpamMatchingEvents
	}
	if from.NumberOfSpamNonMatchingEvents != nil {
		c.NumberOfSpamNonMatchingEvents = from.NumberOfSpamNonMatchingEvents
	}
	if from.CheckBurnAmount != nil {
		c.CheckBurnAmount = from.CheckBurnAmount
	}
	if from.PerformBurnAmount != nil {
		c.PerformBurnAmount = from.PerformBurnAmount
	}
	if from.SharedTrigger != nil {
		c.SharedTrigger = from.SharedTrigger
	}
	if from.UpkeepGasLimit != nil {
		c.UpkeepGasLimit = from.UpkeepGasLimit
	}

	return nil
}

func (c *Load) Validate() error {
	if c.NumberOfUpkeeps == nil || *c.NumberOfUpkeeps < 1 {
		return errors.New("number_of_upkeeps must be set to a positive integer")
	}
	if c.NumberOfEvents == nil || *c.NumberOfEvents < 0 {
		return errors.New("number_of_events must be set to a non-negative integer")
	}
	if c.NumberOfSpamMatchingEvents == nil || *c.NumberOfSpamMatchingEvents < 0 {
		return errors.New("number_of_spam_matching_events must be set to a non-negative integer")
	}
	if c.NumberOfSpamNonMatchingEvents == nil || *c.NumberOfSpamNonMatchingEvents < 0 {
		return errors.New("number_of_spam_non_matching_events must be set to a non-negative integer")
	}
	if c.CheckBurnAmount == nil || c.CheckBurnAmount.Cmp(big.NewInt(0)) < 0 {
		return errors.New("check_burn_amount must be set to a non-negative integer")
	}
	if c.PerformBurnAmount == nil || c.PerformBurnAmount.Cmp(big.NewInt(0)) < 0 {
		return errors.New("perform_burn_amount must be set to a non-negative integer")
	}

	return nil
}
