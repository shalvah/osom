package cmd

import (
	"encoding/json"
	"log/slog"
	config "osom/pkg"
	"strings"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/spf13/cobra"
)

// testRequestCmd represents the testRequest command
var testRequestCmd = &cobra.Command{
	Use:   "test-request",
	Short: "Publishes a test request to the MQTT topic",
	Long:  `Publishes a test request to the MQTT topic to trigger the listen command.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		lat, _ := cmd.Flags().GetString("lat")
		long, _ := cmd.Flags().GetString("long")

		if lat == "" || long == "" {
			parts := strings.Split(config.Config.DefaultLatLong, ",")
			if len(parts) == 2 {
				lat = parts[0]
				long = parts[1]
			} else {
				slog.ErrorContext(ctx, "Invalid DEFAULT_LATLONG format. Expected 'lat,long'.")
				return
			}
		}

		payload := struct {
			Latitude  string `json:"lat"`
			Longitude string `json:"long"`
		}{
			lat, long,
		}

		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to marshal MQTT payload", "error", err)
			return
		}

		mqttClient := mqtt.NewClient(
			mqtt.NewClientOptions().
				AddBroker(config.Config.MQTTUrl).
				SetClientID("osom-tester").
				SetUsername(config.Config.MQTTUsername).
				SetPassword(config.Config.MQTTPassword),
		)
		mqttClient.Connect().Wait()
		token := mqttClient.Publish(config.Config.IncomingRequestsMQTTTopic, 1, false, payloadBytes)
		token.Wait()
		if token.Error() != nil {
			slog.ErrorContext(ctx, "Failed to publish MQTT message", "error", token.Error())
		} else {
			slog.InfoContext(ctx, "Published MQTT message", "topic", config.Config.IncomingRequestsMQTTTopic)
		}
	},
}

func init() {
	rootCmd.AddCommand(testRequestCmd)

	testRequestCmd.Flags().String("lat", "", "Latitude for the request")
	testRequestCmd.Flags().String("long", "", "Longitude for the request")
}
