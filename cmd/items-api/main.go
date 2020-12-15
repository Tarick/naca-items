package main

import (
	"fmt"
	"os"

	"github.com/Tarick/naca-items/internal/logger/zaplogger"

	"github.com/Tarick/naca-items/internal/application/server"
	"github.com/Tarick/naca-items/internal/repository/postgresql"
	"github.com/Tarick/naca-items/internal/version"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func main() {
	var cfgFile string

	// rootCmd represents the base command when called without any subcommands
	rootCmd := &cobra.Command{
		Use:   "items-api",
		Short: "RSS Feeds API",
		Long:  `RSS Feeds API`,
		Run: func(cmd *cobra.Command, args []string) {
			if cfgFile != "" {
				// Use config file from the flag.
				viper.SetConfigFile(cfgFile)
			} else {
				viper.AddConfigPath(".")      // optionally look for config in the working directory
				viper.SetConfigName("config") // name of config file (without extension)
			}
			// If the config file is found, read it in.
			if err := viper.ReadInConfig(); err != nil {
				fmt.Printf("FATAL: error in config file %s. %s", viper.ConfigFileUsed(), err)
				os.Exit(1)
			}

			fmt.Println("Using config file:", viper.ConfigFileUsed())
			// Init logging
			logCfg := &zaplogger.Config{}
			if err := viper.UnmarshalKey("logging", logCfg); err != nil {
				fmt.Println("Failure reading 'logging' configuration:", err)
				os.Exit(1)
			}
			logger := zaplogger.New(logCfg).Sugar()
			defer logger.Sync()

			// Create db configuration
			databaseViperConfig := viper.Sub("database")
			dbCfg := &postgresql.Config{}
			if err := databaseViperConfig.UnmarshalExact(dbCfg); err != nil {
				fmt.Println("FATAL: failure reading 'database' configuration: ", err)
				os.Exit(1)
			}
			// Open db
			itemsRepository, err := postgresql.New(dbCfg, postgresql.NewZapLogger(logger.Desugar()))
			if err != nil {
				fmt.Println("FATAL: failure creating database connection for Items, ", err)
				os.Exit(1)
			}

			// Create web server
			serverCfg := server.Config{}
			serverViperConfig := viper.Sub("server")
			if err := serverViperConfig.UnmarshalExact(&serverCfg); err != nil {
				fmt.Println("FATAL: failure reading 'server' configuration, ", err)
				os.Exit(1)
			}
			httpServer := server.New(serverCfg, logger, itemsRepository)
			httpServer.StartAndServe()
		},
	}
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./config.yaml)")
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version number of application",
		Long:  `Software version`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("NACA RSS Feeds API version:", version.Version, ",build on:", version.BuildTime)
		},
	}
	rootCmd.AddCommand(versionCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
