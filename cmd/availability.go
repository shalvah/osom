/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	config "osom/pkg"
	"osom/pkg/app"
	"strings"
	"time"

	"github.com/spf13/cobra"
)
import mqtt "github.com/eclipse/paho.mqtt.golang"

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
		availability := printAvailability(latLong[0], latLong[1])
		publishAvailability(availability)

		ticker := time.NewTicker(30 * time.Second)
		done := make(chan bool)
		go func() {
			time.Sleep(2 * time.Minute)
			done <- true
		}()
		for {
			select {
			case <-ticker.C:
				printAvailability(latLong[0], latLong[1])
			case <-done:
				ticker.Stop()
				return
			}
		}
	},
}

func publishAvailability(availability []app.LocationAvailability) {
	mqttClient := mqtt.NewClient(
		mqtt.NewClientOptions().
			AddBroker(config.Config.MQTTUrl).
			SetClientID("osom-availability-publisher").
			SetUsername(config.Config.MQTTUsername).
			SetPassword(config.Config.MQTTPassword),
	)
	mqttClient.Connect().Wait()
	fmt.Println("Publishing...")

	for _, loc := range availability {
		payload, _ := json.Marshal(loc)
		mqttClient.Publish(config.Config.AvailabilityMQTTTopic, 1, false, payload).Wait()
		fmt.Println("Published")
	}
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

func printAvailability(lat string, long string) []app.LocationAvailability {
	availability, err := app.FetchAvailability(lat, long)
	if err != nil {
		panic(err)
	}
	fmt.Println(availability)
	return availability
}
