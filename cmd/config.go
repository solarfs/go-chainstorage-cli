package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

type CliConfig struct {
	IpfsGateway      string `toml:"ipfsGateway"`
	GgcscmdPath      string `toml:"ggcscmdPath"`
	UseHTTPSProtocol bool   `toml:"useHttpsProtocol"`
	BucketPrefix     string `toml:"bucketPrefix"`
	ListOffset       int    `toml:"listOffset"`
	CleanTmpData     bool   `toml:"cleanTmpData"`
	MaxRetries       int    `toml:"maxRetries"`
	RetryDelay       int    `toml:"retryDelay"`
}

type SdkConfig struct {
	DefaultRegion            string `toml:"defaultRegion"`
	TimeZone                 string `toml:"timeZone"`
	ChainStorageApiEndpoint  string `toml:"chainStorageApiEndpoint"`
	CarFileWorkPath          string `toml:"carFileWorkPath"`
	CarFileShardingThreshold int    `toml:"carFileShardingThreshold"`
	ChainStorageApiToken     string `toml:"chainStorageApiToken"`
	HttpRequestUserAgent     string `toml:"httpRequestUserAgent"`
	HttpRequestOvertime      int    `toml:"httpRequestOvertime"`
	CarVersion               int    `toml:"carVersion"`
	UseHttpsProtocol         bool   `toml:"useHttpsProtocol"`
}

type LoggerConfig struct {
	LogPath      string `toml:"logPath"`
	Mode         string `toml:"mode"`
	Level        string `toml:"level"`
	IsOutPutFile bool   `toml:"isOutPutFile"`
	MaxAgeDay    int    `toml:"maxAgeDay"`
	RotationTime int    `toml:"rotationTime"`
	UseJSON      bool   `toml:"useJson"`
	LoggerFile   string `toml:"loggerFile"`
}

type CscConfig struct {
	Cli    CliConfig    `toml:"cli"`
	Sdk    SdkConfig    `toml:"sdk"`
	Logger LoggerConfig `toml:"logger"`
}

// region Config show

func configShowRun(cmd *cobra.Command, args []string) {
	configFileUsed := viper.ConfigFileUsed()
	fmt.Fprintln(os.Stderr, "Using config file:", configFileUsed)
	err := printFileContent(configFileUsed)
	if err != nil {
		Error(cmd, args, err)
	}
}

// endregion Config show
