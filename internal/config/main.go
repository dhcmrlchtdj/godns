package config

import (
	"context"
	"encoding/json/v2"
	"os"

	"github.com/rs/zerolog"
)

type Config struct {
	Host     string  `json:"host,omitempty"`
	LogLevel string  `json:"log_level,omitempty"`
	Rule     []*Rule `json:"rule,omitempty"`
	Port     int     `json:"port,omitempty"`
}

type Rule struct {
	Upstream Upstream `json:"upstream"`
	Pattern  Pattern  `json:"pattern"`
}

type Pattern struct {
	Builtin      string   `json:"builtin,omitempty"`
	BuiltinProxy string   `json:"builtin_proxy,omitempty"`
	Record       string   `json:"record,omitempty"`
	Domain       []string `json:"domain,omitempty"`
	Suffix       []string `json:"suffix,omitempty"`
}

type Upstream struct {
	Block    string `json:"block,omitempty"`
	Ipv4     string `json:"ipv4,omitempty"`
	Ipv6     string `json:"ipv6,omitempty"`
	Udp      string `json:"udp,omitempty"`
	Doh      string `json:"doh,omitempty"`
	DohProxy string `json:"doh_proxy,omitempty"`
}

func (c *Config) LoadConfigFile(ctx context.Context, file string) {
	logger := zerolog.Ctx(ctx).
		With().
		Str("module", "config").
		Str("path", file).
		Logger()

	logger.Info().Msg("load config")

	f, err := os.Open(file)
	if err != nil {
		logger.Error().Stack().Err(err).Send()
		panic(err)
	}
	defer f.Close()

	if err := json.UnmarshalRead(f, c, json.StringifyNumbers(true)); err != nil {
		logger.Error().Stack().Err(err).Send()
		panic(err)
	}

	if len(c.Rule) > 0 {
		for _, rule := range c.Rule {
			if err := rule.IsValid(); err != nil {
				logger.Error().Stack().Err(err).Send()
				panic(err)
			}
		}
	}
}
