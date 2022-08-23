package middleware

import (
	"github.com/ory/viper"
	"github.com/ory/x/logrusx"
)

// initializeConfig initializes viper.
func initializeConfig(l *logrusx.Logger, cfgFile string) {
	viper.SetConfigFile(cfgFile)

	// TODO(hperl): Discuss in review (and delete):
	// We probably want to *only* read from the config file in the middleware.
	// viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	// viper.AutomaticEnv() // read in environment variables that match

	err := viper.ReadInConfig()

	if err == nil {
		l.WithField("path", viper.ConfigFileUsed()).Info("Config file loaded successfully.")
	} else {
		switch t := err.(type) {
		case viper.UnsupportedConfigError:
			if len(t) == 0 {
				l.WithError(err).Warn("No config file was defined")
			} else {
				l.WithError(err).WithField("extension", t).Fatal("Unsupported configuration type")
			}
		case *viper.ConfigFileNotFoundError:
			l.WithError(err).Warn("No config file was defined")
		case viper.ConfigFileNotFoundError:
			l.WithError(err).Warn("No config file was defined")
		default:
			l.
				WithField("path", viper.ConfigFileUsed()).
				WithError(err).
				Fatal("Unable to open config file. Make sure it exists and the process has sufficient permissions to read it")
		}
	}
}
