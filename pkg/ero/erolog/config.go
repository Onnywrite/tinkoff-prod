package erolog

import "github.com/Onnywrite/tinkoff-prod/pkg/ero"

type LoggerConfig struct {
	// this is the only option, that won't be updated in Logger.UpdateConfig
	handler string

	addSource      bool
	domainsOptions []LoggerDomainOption
}

type LoggerDomainOption struct {
	domain string
	level  string
}

func NewLoggerDomainOption(domain string, level string) (LoggerDomainOption, ero.Error) {
	faults := make([]LoggerConfigFault, 0)
	if level != "debug" && level != "info" && level != "warn" && level != "error" {
		faults = append(faults, LoggerConfigFault{
			ConfigField: "level",
			Message:     "must be 'debug', 'info', 'warn', or 'error'",
		})
	}

	if len(faults) > 0 {
		return LoggerDomainOption{}, ero.NewValidation(NewContextBuilder().
			WithDomain("erolog.NewLoggerDomainOption").
			With("errors", faults).Build(),
			faults,
		)
	}

	return LoggerDomainOption{
		domain: domain,
		level:  level,
	}, nil
}

type LoggerConfigFault struct {
	ConfigField string
	Message     string
}

func NewConfig(handler string, addSource bool, options ...LoggerDomainOption) (LoggerConfig, ero.Error) {
	var fault *LoggerConfigFault

	if handler != "text" && handler != "json" {
		fault = &LoggerConfigFault{
			ConfigField: "handler",
			Message:     "must be 'text' or 'json'",
		}
	}

	if fault != nil {
		return LoggerConfig{}, ero.NewValidation(NewContextBuilder().
			WithDomain("erolog.NewConfig").
			With("error", *fault).Build(),
			[]LoggerConfigFault{*fault},
		)
	}

	return LoggerConfig{
		handler:        handler,
		addSource:      addSource,
		domainsOptions: options,
	}, nil
}
