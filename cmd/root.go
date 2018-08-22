/*
 * Copyright © 2017-2018 Aeneas Rekkas <aeneas+oss@aeneas.io>
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * @author       Aeneas Rekkas <aeneas+oss@aeneas.io>
 * @copyright  2017-2018 Aeneas Rekkas <aeneas+oss@aeneas.io>
 * @license  	   Apache-2.0
 */

package cmd

import (
	"fmt"
	"os"

	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

var (
	Version   = "dev-master"
	BuildTime = time.Now().String()
	GitHash   = "undefined"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "oathkeeper",
	Short: "A cloud native Access and Identity Proxy",
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

	// RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.oathkeeper.yaml)")
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

	viper.SetDefault("LOG_LEVEL", "info")
	viper.SetDefault("PORT", "4455")
	viper.SetDefault("RULES_REFRESH_INTERVAL", "5s")

	viper.SetDefault("CREDENTIALS_ISSUER_ID_TOKEN_JWK_REFRESH_INTERVAL", "5s")
	viper.SetDefault("CREDENTIALS_ISSUER_ID_TOKEN_HYDRA_JWK_SET_ID", "oathkeeper:id-token")
	viper.SetDefault("CREDENTIALS_ISSUER_ID_TOKEN_ALGORITHM", "HS256")

	viper.SetDefault("AUTHENTICATOR_ANONYMOUS_USERNAME", "anonymous")
	viper.SetDefault("CREDENTIALS_ISSUER_ID_TOKEN_LIFESPAN", "10m")
	viper.SetDefault("CREDENTIALS_ISSUER_ID_TOKEN_ISSUER", "http://localhost:"+viper.GetString("PORT"))

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
