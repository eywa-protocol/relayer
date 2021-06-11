package config

import (
	"time"

	"github.com/spf13/viper"
)

// Config is global object that holds all application level variables.
var Config AppConfig

type AppConfig struct {
	config            string
	TickerInterval    time.Duration
	ECDSA_KEY_2       string
	ECDSA_KEY_1       string
	ECDSA_KEY_3       string
	P2P_PORT          int
	PORT_1            int
	NODELIST_NETWORK3 string
	NODELIST_NETWORK2 string
	NODELIST_NETWORK1 string
	NETWORK_RPC_3     string
	NETWORK_RPC_2     string
	NETWORK_RPC_1     string
	BRIDGE_NETWORK1   string
	BRIDGE_NETWORK2   string
	BRIDGE_NETWORK3   string
}

// LoadConfig loads config from files
func LoadConfig(config AppConfig) error {
	v := viper.New()
	v.SetConfigType("env")
	if config.config == "" {
		v.SetConfigName("bootstrap")
	} else {
		v.SetConfigName(config.config)
	}

	v.SetEnvPrefix("cross-chain")
	v.AutomaticEnv()
	v.AddConfigPath(".")
	err := v.ReadInConfig()
	if err != nil {
		return err
	}
	return v.Unmarshal(&Config)
}

func LoadConfigAndArgs(path string) (err error) {
	cfg := NewConfig(path)
	err = LoadConfig(*cfg)
	return
}

func NewConfig(path string) *AppConfig {
	c := AppConfig{}
	c.config = path
	return &c
}
