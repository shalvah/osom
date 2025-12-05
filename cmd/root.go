package cmd

import (
	"context"
	"os"

	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

var otelShutdown func(context.Context) error
var currentCommandSpan trace.Span
var currentCmdCtx context.Context
var tracer = otel.Tracer("shalvah/osom")

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "osom",
	Short: "Show upcoming departures and bike availability close to you",
	Long:  `Show upcoming departures and bike availability close to yous`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// This is bad for the "listen" command, since it runs like a web server.
		// currentCmdCtx, currentCommandSpan = tracer.Start(cmd.Context(), "command: "+cmd.DisplayName())
		// cmd.SetContext(currentCmdCtx)
		// slog.SetDefault(slog.Default().With(slog.String("trace_id", currentCommandSpan.SpanContext().TraceID().String())))
	},
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func ExecuteContext(ctx context.Context) {
	err := rootCmd.ExecuteContext(ctx)
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.osom.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	// TODO check that this works on SIGTERM
	cobra.OnFinalize(func() {
		// currentCommandSpan.End()
		// ensure we flush before exit
		if otelShutdown != nil {
			_ = otelShutdown(currentCmdCtx)
		}
	})
}
