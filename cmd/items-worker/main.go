package main

import (
	"fmt"
	"os"

	"github.com/Tarick/naca-items/internal/application/worker"
	"github.com/Tarick/naca-items/internal/logger/zaplogger"
	"github.com/Tarick/naca-items/internal/messaging/nsqclient/consumer"
	"github.com/Tarick/naca-items/internal/processor"
	"github.com/Tarick/naca-items/internal/tracing"

	"github.com/Tarick/naca-items/internal/repository/postgresql"
	"github.com/Tarick/naca-items/internal/version"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func main() {
	var (
		cfgFile string
	)
	// rootCmd represents the base command when called without any subcommands
	rootCmd := &cobra.Command{
		Use:   "items-worker",
		Short: "NACA items worker to process news items",
		Long:  `Command line worker for news items processing`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return startWorker(cfgFile)
		},
	}

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version number of application",
		Long:  `Software version`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("NACA Items worker version:", version.Version, "build on:", version.BuildTime)
		},
	}

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./config.yaml)")
	rootCmd.AddCommand(versionCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func startWorker(cfgFile string) error {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath(".")      // optionally look for config in the working directory
		viper.SetConfigName("config") // name of config file (without extension)
	}
	// If the config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("FATAL: error in config file %s. %v", viper.ConfigFileUsed(), err)
	}
	fmt.Println("Using config file:", viper.ConfigFileUsed())
	// Init logging
	logCfg := &zaplogger.Config{}
	if err := viper.UnmarshalKey("logging", logCfg); err != nil {
		return fmt.Errorf("Failure reading 'logging' configuration: %v", err)
	}
	logger := zaplogger.New(logCfg).Sugar()
	defer logger.Sync()

	// Init tracing
	tracingCfg := tracing.Config{}
	if err := viper.UnmarshalKey("tracing", &tracingCfg); err != nil {
		return fmt.Errorf("FATAL: Failure reading 'tracing' configuration, %v", err)
	}
	tracer, tracerCloser, err := tracing.New(tracingCfg, tracing.NewZapLogger(logger))
	defer tracerCloser.Close()
	if err != nil {
		return fmt.Errorf("FATAL: Cannot init tracing, %v", err)
	}

	// Create db configuration
	databaseViperConfig := viper.Sub("database")
	dbCfg := &postgresql.Config{}
	if err := databaseViperConfig.UnmarshalExact(dbCfg); err != nil {
		return fmt.Errorf("FATAL: failure reading 'database' configuration: %v", err)
	}
	// Open db
	repository, err := postgresql.New(dbCfg, postgresql.NewZapLogger(logger.Desugar()), tracer)
	if err != nil {
		return fmt.Errorf("FATAL: failure creating database connection, %v", err)
	}

	consumeViperConfig := viper.Sub("consume")
	consumeCfg := &consumer.MessageConsumerConfig{}
	if err := consumeViperConfig.UnmarshalExact(&consumeCfg); err != nil {
		return fmt.Errorf("FATAL: failure reading 'consume' configuration: %v", err)
	}
	// Construct consumer with message handler
	processor := processor.New(repository, logger, tracer)
	consumer, err := consumer.New(consumeCfg, processor, logger)
	if err != nil {
		return fmt.Errorf("FATAL: consumer creation failed, %v", err)
	}
	wrkr := worker.New(consumer, logger)
	return wrkr.Start()
}
