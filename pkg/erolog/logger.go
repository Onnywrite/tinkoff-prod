package erolog

import (
	"context"
	"io"
	"log/slog"
	"strings"
)

const (
	globalDomain = "all"
)

type secretValue struct {
	value                string
	encryptingPercentage int
}

type Logger struct {
	next     slog.Handler
	levelMap map[string]slog.Level
}

func New(out io.Writer, cfg LoggerConfig) *Logger {
	levelMap := parseDynamicConfig(cfg)

	handlerOptions := &slog.HandlerOptions{
		AddSource: cfg.addSource,
		Level:     slog.LevelDebug,
	}

	var handler slog.Handler
	if cfg.handler == "json" {
		handler = slog.NewJSONHandler(out, handlerOptions)
	} else {
		handler = slog.NewTextHandler(out, handlerOptions)
	}

	return &Logger{
		next:     handler,
		levelMap: levelMap,
	}
}

// Enabled reports whether the handler handles records at the given level.
// The handler ignores records whose level is lower.
func (e *Logger) Enabled(ctx context.Context, level slog.Level) bool {
	if l, ok := getContext(ctx); ok {
		if domainLevel, ok := e.levelMap[l.domain]; ok {
			return domainLevel >= level
		}
	}
	return e.levelMap[globalDomain] <= level
}

// Handle handles the Record.
// It will only be called when Enabled returns true.
// The Context argument is as for Enabled.
// This Handle automatically adds Attrs to a log.
func (e *Logger) Handle(ctx context.Context, rec slog.Record) error {
	if l, ok := getContext(ctx); ok {
		if l.domain != "" {
			rec.Add("domain", l.domain)
		}
		for k, v := range l.attrs {
			if secretVal, ok := v.(secretValue); ok {
				sixtyPercent := len(secretVal.value) / 100 * secretVal.encryptingPercentage
				secretVal.value = strings.Repeat("*", sixtyPercent) + secretVal.value[sixtyPercent:]
				v = secretVal.value
			}
			rec.Add(k, v)
		}
	}
	return e.next.Handle(ctx, rec)
}

// WithAttrs returns a new Handler whose attributes consist of
// both the receiver's attributes and the arguments.
// The Handler owns the slice: it may retain, modify or discard it.
func (e *Logger) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &Logger{
		next:     e.next.WithAttrs(attrs),
		levelMap: e.levelMap,
	}
}

// WithGroup returns a new Handler with the given group appended to
// the receiver's existing groups.
// The keys of all subsequent attributes, whether added by With or in a
// Record, should be qualified by the sequence of group names.
//
// How this qualification happens is up to the Handler, so long as
// this Handler's attribute keys differ from those of another Handler
// with a different sequence of group names.
//
// A Handler should treat WithGroup as starting a Group of Attrs that ends
// at the end of the log event. That is,
//
//	logger.WithGroup("s").LogAttrs(ctx, level, msg, slog.Int("a", 1), slog.Int("b", 2))
//
// should behave like
//
//	logger.LogAttrs(ctx, level, msg, slog.Group("s", slog.Int("a", 1), slog.Int("b", 2)))
//
// If the name is empty, WithGroup returns the receiver.
func (e *Logger) WithGroup(name string) slog.Handler {
	return &Logger{
		next:     e.next.WithGroup(name),
		levelMap: e.levelMap,
	}
}

func (e *Logger) UpdateConfig(newCfg LoggerConfig) {
	e.levelMap = parseDynamicConfig(newCfg)
}

func toSlogLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func parseDynamicConfig(cfg LoggerConfig) map[string]slog.Level {
	levelMap := make(map[string]slog.Level, len(cfg.domainsOptions))
	for _, opts := range cfg.domainsOptions {
		levelMap[opts.domain] = toSlogLevel(opts.level)
	}
	if _, ok := levelMap[globalDomain]; !ok {
		levelMap[globalDomain] = slog.LevelInfo
	}
	return levelMap
}
