/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"encoding/json"
	"log/slog"
	config "osom/pkg"
	"osom/pkg/app"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
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
		availability := printAvailability(cmd.Context(), latLong[0], latLong[1])
		publishAvailability(cmd.Context(), availability)

		ticker := time.NewTicker(30 * time.Second)
		done := make(chan bool)
		go func() {
			time.Sleep(30 * time.Second)
			done <- true
		}()
		for {
			select {
			case <-ticker.C:
				printAvailability(cmd.Context(), latLong[0], latLong[1])
			case <-done:
				ticker.Stop()
				return
			}
		}
	},
}

func publishAvailability(ctx context.Context, availability []app.LocationAvailability) {
	mqttClient := mqtt.NewClient(
		mqtt.NewClientOptions().
			AddBroker(config.Config.MQTTUrl).
			SetClientID("osom-availability-publisher").
			SetUsername(config.Config.MQTTUsername).
			SetPassword(config.Config.MQTTPassword),
	)
	mqttClient.Connect().Wait()

	currentCommandSpan.AddEvent("Published availability to MQTT", trace.WithAttributes(attribute.String("topic", config.Config.AvailabilityMQTTTopic)))
	slog.InfoContext(ctx, "Publishing availability to MQTT", slog.String("topic", config.Config.AvailabilityMQTTTopic))

	payload, _ := json.Marshal(availability)
	mqttClient.Publish(config.Config.AvailabilityMQTTTopic, 1, false, payload).Wait()

	currentCommandSpan.AddEvent("Published availability to MQTT", trace.WithAttributes(attribute.String("topic", config.Config.AvailabilityMQTTTopic)))
	slog.InfoContext(ctx, "Published availability to MQTT")
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

func printAvailability(ctx context.Context, lat string, long string) []app.LocationAvailability {
	currentCommandSpan.AddEvent("Fetching availability from VRN API")
	availability, err := app.FetchAvailability(ctx, lat, long)
	if err != nil {
		panic(err)
	}
	currentCommandSpan.AddEvent("Fetched availability from VRN API")
	return availability
}
