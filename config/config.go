package config

import (
	"github.com/spf13/viper"
	"time"
)

// Config is global object that holds all application level variables.
var Config AppConfig

type AppConfig struct {
	TickerInterval          time.Duration
	ECDSA_KEY_2             string
	ECDSA_KEY_1             string
	P2P_PORT                int
	PORT_1                  int
	BRIDGE_ADDRESS_NETWORK2 string
	BRIDGE_ADDRESS_NETWORK1 string
	NODELIST_NETWORK2       string
	NODELIST_NETWORK1       string
	NETWORK_RPC_2           string
	NETWORK_RPC_1           string
}

// LoadConfig loads config from files
func LoadConfig(config AppConfig) error {
	v := viper.New()
	v.SetConfigType("env")
	v.SetConfigName("bootstrap")
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
	return &c
}
