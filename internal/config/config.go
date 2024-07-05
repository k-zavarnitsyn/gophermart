package config

import (
	"crypto/ecdsa"
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/caarlos0/env/v10"
	"github.com/golang-jwt/jwt/v5"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

const (
	DefaultDir  = "config"
	DefaultFile = "config.yaml"
	LocalFile   = "local.yaml"
	DefaultDSN  = ""
)

var Default = &Config{
	Host:    "localhost",
	Port:    8080,
	Scheme:  "http",
	Address: "localhost:8080",

	Log: Log{
		Level:            log.InfoLevel,
		WithReportCaller: false,
		WithRequestData:  true,
		WithResponseData: true,
	},
	Server: Server{
		DatabaseURI:       DefaultDSN,
		ReadHeaderTimeout: 3 * time.Second,
		ShutdownTimeout:   10 * time.Second,
	},
	Auth: Auth{
		PasswordHashKey: []byte("761b13f9e49816b818cc317f73727bbd3cfc23fa"),
		ExpiresIn:       24 * 7 * time.Hour,
		Leeway:          time.Second * 10,
		ValidMethods:    []string{"ES256"},
		CookieName:      "access_token",
	},
	Accrual: Accrual{
		MaxActiveWorkers:    100,
		OverloadReportCount: 1000,
		OverloadReportRPS:   50,
		PollingInterval:     time.Second,
		PollingCount:        1000,
	},
}

// generation tool: https://zhwt.github.io/yaml-to-go/

type Option func(config *Config) error

type Config struct {
	Host    string `yaml:"host" env:"HOST"`
	Port    int    `yaml:"port" env:"PORT"`
	Scheme  string `yaml:"scheme"`
	Address string `env:"RUN_ADDRESS,expand"`

	Log     Log     `yaml:"log"`
	Server  Server  `yaml:"server"`
	Auth    Auth    `yaml:"auth"`
	Accrual Accrual `yaml:"accrual"`

	baseDir string
}

type Log struct {
	Level            log.Level
	WithReportCaller bool
	WithRequestData  bool
	WithResponseData bool
}

type Server struct {
	DatabaseURI       string        `yaml:"databaseURI" env:"DATABASE_URI"`
	ReadHeaderTimeout time.Duration `yaml:"readHeaderTimeout"`
	ShutdownTimeout   time.Duration `yaml:"shutdownTimeout"`
}

type Auth struct {
	PasswordHashKey []byte        `yaml:"passwordHashKey" env:"PASSWORD_HASH_KEY"`
	PemKeyFile      string        `yaml:"pemKeyFile" env:"PEM_KEY_FILE"`
	ExpiresIn       time.Duration `yaml:"expiresIn" env:"EXPIRES_IN"`
	JwtPrivateKey   *ecdsa.PrivateKey
	Leeway          time.Duration
	ValidMethods    []string
	CookieName      string
}

type Accrual struct {
	AccrualSystemAddress string        `yaml:"accrualSystemAddress" env:"ACCRUAL_SYSTEM_ADDRESS"`
	MaxActiveWorkers     int           `yaml:"maxActiveWorkers" env:"MAX_ACTIVE_WORKERS"`
	OverloadReportCount  int           `yaml:"overloadReportCount"`
	OverloadReportRPS    float64       `yaml:"overloadReportRps"`
	PollingInterval      time.Duration `yaml:"pollingInterval"`
	PollingCount         int           `yaml:"pollingCount"`
}

func LoadYaml(dir string) (*Config, error) {
	fileData, err := os.ReadFile(dir + "/" + LocalFile)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		fileData, err = os.ReadFile(dir + "/" + DefaultFile)
	}
	if err != nil {
		return nil, err
	}

	cfg := Default
	if err := yaml.Unmarshal(fileData, cfg); err != nil {
		return nil, err
	}
	cfg.Address = fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	cfg.baseDir = strings.TrimRight(dir, "/")

	return cfg, nil
}

func (c *Config) WithOptions(opts ...Option) error {
	for _, opt := range opts {
		if err := opt(c); err != nil {
			return err
		}
	}

	return nil
}

func Load(dir string, opts ...Option) (*Config, error) {
	cfg, err := LoadYaml(dir)
	if err != nil {
		return nil, err
	}
	cfg.addCommonFlags()
	if err := cfg.WithOptions(opts...); err != nil {
		return nil, err
	}
	flag.Parse()
	err = cfg.ParseEnv()

	return cfg, err
}

func (c *Config) ParseEnv() error {
	opts := env.Options{
		OnSet: func(tag string, value interface{}, isDefault bool) {
			if tag == "RUN_ADDRESS" && !isDefault && value != "" {
				host, port, err := ParseServerHostPort(value.(string))
				if err != nil {
					panic(fmt.Errorf("unable to load value from RUN_ADDRESS env: %w", err))
				}
				c.Host, c.Port = host, port
			}
		},
	}
	if err := env.ParseWithOptions(c, opts); err != nil {
		return err
	}
	return nil
}

func (c *Config) addCommonFlags() {
	flag.Func("a", "HTTP-server endpoint address (host:port)", func(flagValue string) error {
		host, port, err := ParseServerHostPort(flagValue)
		if err != nil {
			return err
		}
		c.Host, c.Port = host, port
		c.Address = flagValue

		return nil
	})
	flag.Func("r", "Accrual system address", func(flagValue string) error {
		c.Accrual.AccrualSystemAddress = flagValue

		return nil
	})
}

// ParseServerHostPort парсит хост и порт, нужна т.к. url.Parse парсит лишнее
func ParseServerHostPort(address string) (host string, port int, err error) {
	addr := strings.Split(address, ":")
	if len(addr) > 2 {
		return "", 0, errors.New("need address in a form host:port")
	}
	port = 80
	if len(addr) == 2 {
		port, err = strconv.Atoi(addr[1])
		if err != nil {
			return "", 0, fmt.Errorf("unable to determine port: %w", err)
		}
	}
	return addr[0], port, nil
}

func (c *Config) UseDB() bool {
	return c.Server.DatabaseURI != ""
}

func WithServerFlags() Option {
	return func(c *Config) error {
		flag.Func("d", fmt.Sprintf("Database URI [default:%s]", DefaultDSN), func(flagValue string) error {
			c.Server.DatabaseURI = flagValue

			return nil
		})

		return nil
	}
}

func WithAuth() Option {
	return func(cfg *Config) error {
		// for developing and testing
		if cfg.Auth.PemKeyFile == "" {
			cfg.Auth.PemKeyFile = cfg.baseDir + "/private/default-key.pem"
		}
		pemData, err := os.ReadFile(cfg.Auth.PemKeyFile)
		if err != nil {
			return fmt.Errorf("unable to read pem file: %v", err)
		}
		cfg.Auth.JwtPrivateKey, err = jwt.ParseECPrivateKeyFromPEM(pemData)
		if err != nil {
			return fmt.Errorf("unable to parse pem file: %v", err)
		}

		return nil
	}
}
