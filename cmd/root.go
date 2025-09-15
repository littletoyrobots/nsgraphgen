/*
Copyright © 2025 Adam Yarborough @littletoyrobots
*/
package cmd

import (
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"slices"

	"os"
	"strings"

	"github.com/littletoyrobots/nsgraphgen/internal/graphgen"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// used for flags

var cfgFile string

// var rankDir string
// var inputFile string
// var outputFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "nsgraphgen",
	Short:   "Generate graphs from Netscaler configs to create graphviz dot and mermaid compatible outputs.",
	Version: "0.1.0",
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error { return initializeConfig(cmd) },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		log.Fatal(err)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().String("rankdir", "TB", "graph rank direction")
	rootCmd.PersistentFlags().StringP("input-file", "i", "ns.conf", "input netscaler config file")
	rootCmd.PersistentFlags().StringP("output-file", "o", "graph.out", "output graph to file")
	rootCmd.PersistentFlags().StringSlice("ignore-name", []string{}, "names of resources to ignore from graphs")
	rootCmd.PersistentFlags().StringSlice("ignore-type", []string{}, "names of types to ignore from graphs")
	rootCmd.PersistentFlags().StringSlice("isolate-name", []string{}, "names of resources to isolate in graph")
	rootCmd.PersistentFlags().Bool("stdout", false, "output to STDOUT, overrides output-file")

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "nsgraphgen config file (default: ./config.yaml)")
	rootCmd.PersistentFlags().BoolP("verbose", "V", false, "enable verbose output")
	rootCmd.PersistentFlags().BoolP("quiet", "q", false, "suppress all log output")
	rootCmd.SilenceErrors = true
	// rootCmd.SilenceUsage = true
	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func initializeConfig(cmd *cobra.Command) error {
	// cmd.SilenceUsage = true
	// setup viper to use environmental variables
	viper.SetEnvPrefix("NSGRAPHGEN")
	// allow for nested keys in environmental variables, e.g. `NSGRAPHGEN_DB_HOST` or whatever
	// viper.SetEnvKeyReplacer(strings.NewReplacer(".", "*", "-", "*"))
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		// Search for a config file in default locations.
		home, err := os.UserHomeDir()
		// Only panic if we can't get the home directory.
		cobra.CheckErr(err)

		// Search for a config file with the name "config" (without extension).
		viper.AddConfigPath(".")
		viper.AddConfigPath(home + "/.config/nsgraphgen")
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}

	// 3. Read the configuration file.
	// If a config file is found, read it in. We use a robust error check
	// to ignore "file not found" errors, but panic on any other error.
	if err := viper.ReadInConfig(); err != nil {
		// It's okay if the config file doesn't exist.
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			return err
		}
	}

	// 4. Bind Cobra flags to Viper.
	// This is the magic that makes the flag values available through Viper.
	// It binds the full flag set of the command passed in.
	err := viper.BindPFlags(cmd.Flags())
	if err != nil {
		return err
	}

	verbose := viper.GetBool("verbose")
	quiet := viper.GetBool("quiet")

	for _, each := range viper.GetStringSlice("ignore-type") {
		if !slices.Contains(graphgen.NodeTypes[:], each) {
			return fmt.Errorf("invalid ignore-type: %v. \nvalue must be in %v", each, graphgen.NodeTypes)
		}
	}

	if quiet {
		log.SetOutput(io.Discard)
	}
	if verbose {
		slog.SetLogLoggerLevel(slog.LevelDebug)
		slog.Info("starting nsgraphgen")
		slog.Debug("verbose output enabled")
	}

	return nil
}
