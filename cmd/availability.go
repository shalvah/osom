/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	config "osom/pkg"
	"osom/pkg/app"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// availabilityCmd represents the availability command
var availabilityCmd = &cobra.Command{
	Use:   "availability",
	Short: "Show available bikes",
	Long:  ``,
	PreRun: func(cmd *cobra.Command, args []string) {
		app.GetUserLocation()
	},
	Run: func(cmd *cobra.Command, args []string) {
		latLong := strings.Split(config.Config.DefaultLatLong, ",")
		printAvailability(latLong[0], latLong[1])

		ticker := time.NewTicker(30 * time.Second)
		for {
			<-ticker.C
			printAvailability(latLong[0], latLong[1])
		}
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

func printAvailability(lat string, long string) {
	availability, err := app.FetchAvailability(lat, long)
	if err != nil {
		panic(err)
	}
	fmt.Println(availability)
}
