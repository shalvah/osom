package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	config "osom/pkg"
	"osom/pkg/app"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var activeSpans []trace.Span

// listenCmd represents the listen command
var listenCmd = &cobra.Command{
	Use:   "listen",
	Short: "Serves MQTT requests for availability and departures",
	Long:  `Listens for incoming MQTT requests and publishes availability and departure information.`,
	PreRun: func(cmd *cobra.Command, args []string) {
	},
	Run: func(cmd *cobra.Command, args []string) {
		parentCtx := cmd.Context()
		ctx := context.WithValue(parentCtx, "rootCmd", cmd)
		// mqtt.DEBUG = log.New(os.Stdout, "[DEBUG] ", 0)
		mqttClient := mqtt.NewClient(
			mqtt.NewClientOptions().
				AddBroker(config.Config.MQTTUrl).
				SetClientID("osom-server").
				SetUsername(config.Config.MQTTUsername).
				SetPassword(config.Config.MQTTPassword).
				SetOnConnectHandler(func(c mqtt.Client) {
					listenForRequests(ctx, c)
					listenForErrors(ctx, c)
				}).SetConnectionLostHandler(mqtt.DefaultConnectionLostHandler),
		)
		mqttClient.Connect().Wait()

		// Block until context is cancelled (SIGINT, SIGTERM etc), so our subscriber does not exit
		<-parentCtx.Done()
		// End all spans in progress
		for _, span := range activeSpans {
			span.End()
		}
	},
}

func init() {
	rootCmd.AddCommand(listenCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listenCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listenCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

type MQTTRequestPayload struct {
	Latitude  string `json:"lat"`
	Longitude string `json:"long"`
}

func listenForRequests(ctx context.Context, mqttClient mqtt.Client) {
	subscribeToken := mqttClient.Subscribe(config.Config.IncomingRequestsMQTTTopic, 1, func(client mqtt.Client, msg mqtt.Message) {
		// Create a new context, so each MQTT request is traced separately.
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		ctx, span := otel.Tracer("shalvah/osom").Start(ctx, "handle MQTT request")
		activeSpans = append(activeSpans, span)
		defer func() {
			span.End()
			for i, s := range activeSpans {
				if s == span {
					activeSpans = append(activeSpans[:i], activeSpans[i+1:]...)
					break
				}
			}
		}()

		slog.SetDefault(slog.Default().With(slog.String("trace_id", span.SpanContext().TraceID().String())))

		slog.InfoContext(ctx, "Received MQTT request", slog.String("topic", msg.Topic()))
		payloadStr := msg.Payload()
		span.AddEvent("Received MQTT request", trace.WithAttributes(
			attribute.String("topic", msg.Topic()),
			attribute.String("payload", string(payloadStr)),
		))

		var payloadParsed MQTTRequestPayload
		err := json.Unmarshal(payloadStr, &payloadParsed)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to parse MQTT request payload", slog.String("error", err.Error()))
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return
		}

		availability, err := app.FetchAvailability(ctx, payloadParsed.Latitude, payloadParsed.Longitude)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return
		}
		publishAvailability(ctx, availability)
	})

	subscribeToken.Wait()
	if err := subscribeToken.Error(); err != nil {
		panic(fmt.Sprintf("Failed to subscribe to MQTT topic: %s", err.Error()))
	} else {
		slog.InfoContext(ctx, "Subscribed to MQTT topic", slog.String("topic", config.Config.IncomingRequestsMQTTTopic))
	}
}

func listenForErrors(ctx context.Context, mqttClient mqtt.Client) {
	subscribeToken := mqttClient.Subscribe("shalvah/errors", 1, func(client mqtt.Client, msg mqtt.Message) {
		payloadStr := msg.Payload()
		slog.ErrorContext(ctx, "Received MQTT error", slog.String("payload", string(payloadStr)))
	})

	subscribeToken.Wait()
	if err := subscribeToken.Error(); err != nil {
		panic(fmt.Sprintf("Failed to subscribe to MQTT errors topic: %s", err.Error()))
	} else {
		slog.InfoContext(ctx, "Subscribed to MQTT errors topic")
	}
}
