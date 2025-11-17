/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"osom/pkg/app"

	"github.com/spf13/cobra"
)

// departuresCmd represents the check command
var departuresCmd = &cobra.Command{
	Use:   "departures",
	Short: "Find upcoming departures",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		departures := app.FetchDepartures()
		fmt.Println(departures)
	},
}

func init() {
	rootCmd.AddCommand(departuresCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// departuresCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// departuresCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
