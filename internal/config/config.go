package config

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

var (
	ConfigPathFlag = "config-path"
	ConfigPathEnv  = "CONFIG_PATH"
)

type Config struct {
	Conn        string        `yaml:"conn"`
	WatchFreq   time.Duration `yaml:"watch_freq"`
	ServiceName string        `yaml:"service_name"`

	Https        TransportConfig `yaml:"https"`
	AccessToken  TokenConfig     `yaml:"access_token" dynamic:"true"`
	RefreshToken TokenConfig     `yaml:"refresh_token" dynamic:"true"`

	Logger LoggerConfig `yaml:"logger"`

	path string
	dir  string
	t    *time.Ticker
}

type TransportConfig struct {
	Port uint16 `yaml:"port"`
	Cert string `yaml:"cert"`
	Key  string `yaml:"key"`
}

type TokenConfig struct {
	Secret   string        `yaml:"secret" dynamic:"true"`
	TTL      time.Duration `yaml:"ttl" dynamic:"true"`
	Issuer   string        `yaml:"issuer"`
	Audience string        `yaml:"audience"`
	Subject  string        `yaml:"subject"`
}

type LoggerConfig struct {
	Handler        string               `yaml:"handler"`
	Out            string               `yaml:"out"`
	AddSource      bool                 `yaml:"add_source"`
	DomainsOptions []LoggerDomainOption `yaml:"domains_options" dynamic:"true"`
}

type LoggerDomainOption struct {
	Domain string `yaml:"domain"`
	Level  string `yaml:"level"`
}

func MustLoad(defaultPath string) *Config {
	conf, err := Load(defaultPath)
	if err != nil {
		panic(err)
	}
	return conf
}

func Load(defaultPath string) (*Config, error) {
	var configPath string
	flag.StringVar(&configPath, ConfigPathFlag, "", "config file path")
	flag.Parse()

	if configPath == "" {
		configPath = os.Getenv(ConfigPathEnv)
	}

	if configPath == "" {
		configPath = defaultPath
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("%w: path '%s'", err, configPath)
	}
	return LoadPath(configPath)
}

func LoadPath(path string) (*Config, error) {
	var cfg Config
	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		return nil, fmt.Errorf("config could not be loaded: %w", err)
	}
	cfg.path = path

	folders := strings.Split(cfg.path, "/")
	cfg.dir = strings.Join(folders[:len(folders)-1], "/")

	return &cfg, nil
}

func (c *Config) Dir() string {
	return c.dir
}

func (c *Config) ResetWatchFreq(freq time.Duration) {
	if c.WatchFreq == freq {
		return
	}
	if freq > 0 {
		c.WatchFreq = freq
		c.t.Reset(c.WatchFreq)
	}
}

func (c *Config) StartWatch(ctx context.Context, onChange func(Config)) {
	c.t = time.NewTicker(c.WatchFreq)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-c.t.C:
				c.watch(onChange)
			}
		}
	}()
}

func (c *Config) watch(callback func(Config)) {
	if newPath, changed := updatePath(c.path); changed {
		c.path = newPath
	}
	info, err := os.Stat(c.path)
	if os.IsNotExist(err) {
		return
	}
	if info.ModTime().Before(time.Now().Add(-c.WatchFreq)) {
		return
	}

	newcfg, err := LoadPath(c.path)
	if err != nil {
		return
	}
	callback(*newcfg)
}

func updatePath(oldPath string) (string, bool) {
	newPath := os.Getenv(ConfigPathEnv)
	if newPath == "" || newPath == oldPath {
		return "", false
	}
	return newPath, true
}
