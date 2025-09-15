/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	//	"fmt"

	"log"

	"github.com/littletoyrobots/nsgraphgen/internal/graphgen"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// dotCmd represents the dot command
var dotCmd = &cobra.Command{
	Use:     "dot",
	Short:   "Generate a graph in dot format",
	Aliases: []string{"graphviz"},
	RunE: func(cmd *cobra.Command, args []string) error {
		rankdir := viper.GetString("rankdir")
		inputFile := viper.GetString("input-file")
		outputFile := viper.GetString("output-file")
		ignoreNames := viper.GetStringSlice("ignore-name")
		ignoreTypes := viper.GetStringSlice("ignore-type")
		isolateNames := viper.GetStringSlice("isolate-name")
		stdout := viper.GetBool("stdout")

		ns := graphgen.New(rankdir, ignoreNames, ignoreTypes, isolateNames)
		if err := ns.Parse(inputFile); err != nil {
			log.Fatal(err)
		}
		ns.ExportDot(outputFile, stdout)
		return nil
	},
	//Run: func(cmd *cobra.Command, args []string) {
	//	fmt.Println("dot called")
	//},
}

func init() {

	rootCmd.AddCommand(dotCmd)

	// dotCmd.Flags().Bool("dark-mode", false, "set output to dark mode")
	// dotCmd.Flags().Bool("include-legend", false, "include legend / key")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// dotCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// dotCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
