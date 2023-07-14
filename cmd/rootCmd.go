/*
Copyright © 2023 pld

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"fmt"
	chainstoragesdk "github.com/solarfs/go-chainstorage-sdk"
	"github.com/ulule/deepcopier"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var sdkCfgFile string
var debug bool
var cliConfig CliConfig
var sdkConfig SdkConfig
var loggerConfig LoggerConfig
var appConfig chainstoragesdk.ApplicationConfig

type PlainFormatter struct {
}

func (f *PlainFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	return []byte(fmt.Sprintf("%s\n", entry.Message)), nil
}

func toggleDebug(cmd *cobra.Command, args []string) {
	if debug {
		logrus.Info("Debug logs enabled")
		logrus.SetLevel(logrus.DebugLevel)
		logrus.SetFormatter(&logrus.TextFormatter{})
	} else {
		plainFormatter := new(PlainFormatter)
		logrus.SetFormatter(plainFormatter)
	}
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gcscmd [--config=<config> | -c] [--debug | -D] [--help] [-h]  [--timeout=<timeout>] <command> ...",
	Short: "Golang ChainStorage Command line tool",
	Long:  `Golang ChainStorage Command line tool`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
	//PreRun: toggleDebug,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	//rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./config.toml)")
	//rootCmd.PersistentFlags().StringVar(&sdkCfgFile, "sdkConfig", "", "sdk config file (default is ./chainstorage-sdk.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	//rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	//rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", true, "verbose logging")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		//home, err := os.UserHomeDir()
		//cobra.CheckErr(err)

		// Search config in home directory with name ".chainstorage-cli" (without extension).
		//viper.AddConfigPath(home)
		//viper.SetConfigType("yaml")
		//viper.SetConfigName(".chainstorage-cli")

		viper.SetConfigName("config")
		viper.SetConfigType("toml")
		viper.AddConfigPath(".")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		//fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())

		// todo: exit cli, if config can't be found?
		fmt.Fprintf(os.Stderr, "config file can't be found, error:%v\n", err)
		os.Exit(1)
	}

	//cscConfig := CscConfig{}
	//err := viper.UnmarshalExact(&cscConfig)
	//if err != nil {
	//	fmt.Fprintln(os.Stderr, "viper.UnmarshalExact, error:", err)
	//}

	cscConfig := CscConfig{}
	err := viper.Unmarshal(&cscConfig)
	if err != nil {
		fmt.Fprintln(os.Stderr, "config Unmarshal fail, error:%+v\n", err)
		os.Exit(1)
	}

	//fmt.Printf("Config struct: %#v\n", cscConfig)
	//fmt.Printf("All configuration: %+v\n", viper.AllSettings())
	appConfig = chainstoragesdk.ApplicationConfig{}

	checkConfig(&cscConfig)

	cliConfig = cscConfig.Cli
	sdkConfig = cscConfig.Sdk
	loggerConfig = cscConfig.Logger

	// 设置SDK配置
	deepcopier.Copy(&sdkConfig).To(&appConfig.Server)
	deepcopier.Copy(&loggerConfig).To(&appConfig.Logger)

	initLogger()
}

func checkConfig(config *CscConfig) {
	cliConfig := &config.Cli

	//check chain-storage-api base address
	if len(cliConfig.IpfsGateway) > 0 {
		ipfsGateway := cliConfig.IpfsGateway
		if !strings.HasPrefix(ipfsGateway, "http://") &&
			!strings.HasPrefix(ipfsGateway, "https://") {

			if cliConfig.UseHTTPSProtocol {
				cliConfig.IpfsGateway = "https://" + ipfsGateway
			} else {
				cliConfig.IpfsGateway = "http://" + ipfsGateway
			}
		}

		if !strings.HasSuffix(ipfsGateway, "/") {
			cliConfig.IpfsGateway += "/"
		}
	} else {
		fmt.Println("ERROR: no ipfs gateway provided in Configuration, at least 1 valid http/https ipfs gateway must be given, exiting")
		os.Exit(1)
	}

}
