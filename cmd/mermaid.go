/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	//	"fmt"
	"github.com/littletoyrobots/nsgraphgen/internal/graphgen"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// mermaidCmd represents the mermaid command
var mermaidCmd = &cobra.Command{
	Use:   "mermaid",
	Short: "Generate a graph in mermaid format",
	RunE: func(cmd *cobra.Command, args []string) error {
		rankdir := viper.GetString("rankdir")
		inputFile := viper.GetString("input-file")
		outputFile := viper.GetString("output-file")
		ignoreNames := viper.GetStringSlice("ignore-name")
		ignoreTypes := viper.GetStringSlice("ignore-type")
		isolateNames := viper.GetStringSlice("isolate-name")
		stdout := viper.GetBool("stdout")

		ns := graphgen.New(rankdir, ignoreNames, ignoreTypes, isolateNames)
		ns.Parse(inputFile)
		ns.ExportMermaid(outputFile, stdout)
		return nil
	},
}

func init() {
	// mermaidCmd.SilenceUsage = true
	rootCmd.AddCommand(mermaidCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// mermaidCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// mermaidCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
