package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/sirupsen/logrus"
)

var cfgFile string

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "oathkeeper",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

var logger *logrus.Logger

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports Persistent Flags, which, if defined here,
	// will be global for your application.

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.oathkeeper.yaml)")
	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	RootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
	}

	viper.SetConfigName(".oathkeeper") // name of config file (without extension)
	viper.AddConfigPath("$HOME")       // adding home directory as first search path
	viper.AutomaticEnv()               // read in environment variables that match

	viper.BindEnv("LOG_LEVEL")
	viper.SetDefault("LOG_LEVEL", "info")

	viper.BindEnv("DATABASE_URL")
	viper.SetDefault("DATABASE_URL", "")

	viper.BindEnv("HYDRA_CLIENT_ID")
	viper.SetDefault("HYDRA_CLIENT_ID", "")

	viper.BindEnv("HYDRA_CLIENT_SECRET")
	viper.SetDefault("HYDRA_CLIENT_SECRET", "")

	viper.BindEnv("HYDRA_URL")
	viper.SetDefault("HYDRA_URL", "")

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}

	logLevel, err := logrus.ParseLevel(viper.GetString("LOG_LEVEL"))
	if err != nil {
		logLevel = logrus.InfoLevel
	}

	logger = logrus.New()
	logger.Level = logLevel
}
