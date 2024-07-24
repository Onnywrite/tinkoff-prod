package config

import (
	"github.com/Onnywrite/tinkoff-prod/pkg/ero"
	"github.com/Onnywrite/tinkoff-prod/pkg/erolog"
)

func (c *Config) MustErologConfig() erolog.LoggerConfig {
	cfg, err := c.ErologConfig()
	if err != nil {
		panic(err)
	}
	return cfg
}

func (c *Config) ErologConfig() (erolog.LoggerConfig, ero.Error) {
	opts := make([]erolog.LoggerDomainOption, 0, len(c.Logger.DomainsOptions))

	for _, opt := range c.Logger.DomainsOptions {
		erologOpt, eroErr := erolog.NewLoggerDomainOption(opt.Domain, opt.Level)
		if eroErr != nil {
			return erolog.LoggerConfig{}, eroErr
		}
		opts = append(opts, erologOpt)
	}

	cfg, eroErr := erolog.NewConfig(c.Logger.Handler, c.Logger.AddSource, opts...)
	if eroErr != nil {
		return erolog.LoggerConfig{}, eroErr
	}

	return cfg, nil
}
