package config

import (
	"flag"
	//"context"
	"fmt"
	"github.com/spf13/viper"
	"time"
)

// Config is global object that holds all application level variables.
var Config AppConfig

type AppConfig struct {
	CONFIG_PATH string

	TickerInterval time.Duration

	ECDSA_KEY_2 string
	ECDSA_KEY_1 string

	P2P_PORT int

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
}

func ReadConfig() error {
	viper.SetConfigName("env_p2p_bridge") // name of config file (without extension)
	viper.SetEnvPrefix("cross-chain")
	viper.SetConfigType("env")   // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath("$HOME") // call multiple times to add many search paths
	viper.AddConfigPath(".")     // optionally look for config in the working directory
	err := viper.ReadInConfig()  // Find and read the config file
	if err != nil {              // Handle errors reading the config file
		//panic(fmt.Errorf("Fatal error config file: %s \n", err))
		return err
	}
	return nil
}

// LoadConfig loads config from files
func LoadConfig(configPaths ...string) error {
	v := viper.New()
	v.SetConfigName("env_p2p_bridge.env")
	v.SetConfigType("env")
	v.SetEnvPrefix("cross-chain")
	v.AutomaticEnv()
	for _, path := range configPaths {
		v.AddConfigPath(path)
		fmt.Print("path", path, "\n")
	}
	err := v.ReadInConfig()
	if err != nil {
		return err
	}

	return v.Unmarshal(&Config)
}

func LoadConfigAndArgs() (cfg *AppConfig, err error) {
	cfg = NewConfig()
	err = LoadConfig(cfg.CONFIG_PATH)
	return
}

func NewConfig() *AppConfig {
	c := AppConfig{}
	flag.StringVar(&c.CONFIG_PATH, "cnf", ".", "config path absolute path")
	flag.Parse()
	return &c
}
