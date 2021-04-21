package config

import (
	"flag"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
)

// Config is global object that holds all application level variables.
var Config AppConfig

type AppConfig struct {
	CONFIG_PATH               string
	CONFIG_NAME               string
	TickerInterval            time.Duration
	ECDSA_KEY_2               string
	ECDSA_KEY_1               string
	P2P_PORT                  int
	PORT_2                    int
	PORT_1                    int
	LISTEN_NETWORK_2          string
	LISTEN_NETWORK_1          string
	ORACLE_CONTRACT_ADDRESS_2 string
	ORACLE_CONTRACT_ADDRESS_1 string
	TOKENPOOL_ADDRESS_2       string
	TOKENPOOL_ADDRESS_1       string
	BRIDGE_ADDRESS_2          string
	BRIDGE_ADDRESS_1          string
	NETWORK_RPC_2             string
	NETWORK_RPC_1             string
	KEY_FILE                  string
	BOOTSTRAP_PEER            string
}

// LoadConfig loads config from files
func LoadConfig(config AppConfig) error {
	v := viper.New()
	v.SetConfigType("env")
	v.SetConfigName(config.CONFIG_NAME)
	v.SetEnvPrefix("cross-chain")
	v.AutomaticEnv()
	v.AddConfigPath(config.CONFIG_PATH)
	err := v.ReadInConfig()
	if err != nil {
		return err
	}
	return v.Unmarshal(&Config)
}

func LoadConfigAndArgs() (cfg *AppConfig, err error) {
	cfg = NewConfig()
	err = LoadConfig(*cfg)
	return
}

func NewConfig() *AppConfig {
	c := AppConfig{}
	var path string
	flag.StringVar(&path, "cnf", ".", "config file absolute path")
	flag.Parse()
	c.CONFIG_PATH = filepath.Dir(path)
	c.CONFIG_NAME = filepath.Base(path)
	return &c
}
