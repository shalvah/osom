package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	config "osom/pkg"
	"osom/pkg/app"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/getsentry/sentry-go"
	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var activeSpans []trace.Span

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
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
					listenForErrors(ctx, c)
					listenForRequests(ctx, c)
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
	rootCmd.AddCommand(serverCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// serverCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// serverCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

type MQTTRequestPayload struct {
	Latitude  string `json:"lat"`
	Longitude string `json:"long"`
}

func listenForRequests(ctx context.Context, mqttClient mqtt.Client) {
	subscribeToken := mqttClient.Subscribe(config.Config.IncomingRequestsMQTTTopic, 1, func(client mqtt.Client, msg mqtt.Message) {
		// Create a new Sentry and OTel context, so each MQTT request is traced separately.
		// hub := sentry.NewHub(sentry.CurrentHub().Client(), sentry.NewScope())
		// hub.ConfigureScope(func(scope *sentry.Scope) {
		// 	scope.SetContext("mqtt", map[string]interface{}{
		// 		"topic":   msg.Topic(),
		// 		"payload": string(msg.Payload()),
		// 	})
		// })
		// ctx = sentry.SetHubOnContext(ctx, hub)
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
		messageBytes := msg.Payload()
		span.AddEvent("Received MQTT request", trace.WithAttributes(
			attribute.String("topic", msg.Topic()),
			attribute.String("payload", string(messageBytes)),
		))

		var payloadParsed MQTTRequestPayload
		err := parseMqttMessage(ctx, messageBytes, &payloadParsed)
		if err != nil {
			return
		}

		availability, err := app.FetchAvailability(ctx, payloadParsed.Latitude, payloadParsed.Longitude)
		if err != nil {
			recordError(ctx, err, "Failed to fetch availability")
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

// Adafruit communicates publish/subscribe errors on error topics.
// See https://io.adafruit.com/api/docs/mqtt.html#troubleshooting
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

func parseMqttMessage[T any](ctx context.Context, messageBytes []byte, payloadParsed *T) error {
	var messageParsed struct {
		Data struct {
			Value string `json:"value"`
		} `json:"data"`
	}
	err := json.Unmarshal(messageBytes, &messageParsed)
	if err != nil {
		recordError(ctx, err, "Failed to parse MQTT message wrapper")
		return err
	}

	err = json.Unmarshal([]byte(messageParsed.Data.Value), payloadParsed)
	if err != nil {
		recordError(ctx, err, "Failed to parse MQTT request payload")
		return err
	}
	return nil
}

func recordError(ctx context.Context, err error, msg string) {
	span := trace.SpanFromContext(ctx)
	// hub := sentry.GetHubFromContext(ctx)
	// hub.Scope().SetContext("trace", map[string]interface{}{
	// 	"trace_id": span.SpanContext().TraceID().String(),
	// 	"span_id":  span.SpanContext().SpanID().String(),
	// })
	// hub.CaptureException(err)
	sentry.CaptureException(err)
	slog.ErrorContext(ctx, msg, slog.String("error", err.Error()))
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
}
