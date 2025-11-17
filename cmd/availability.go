/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"osom/pkg/app"

	"github.com/spf13/cobra"
)

// availabilityCmd represents the availability command
var availabilityCmd = &cobra.Command{
	Use:   "availability",
	Short: "Show available bikes",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		availability, err := app.FetchAvailability()
		if err != nil {
			panic(err)
		}
		fmt.Println(availability)
	},
}

func init() {
	rootCmd.AddCommand(availabilityCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// availabilityCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// availabilityCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
