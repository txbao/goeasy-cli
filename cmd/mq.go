package cmd

import (
	"path/filepath"

	"github.com/txbao/goeasy-cli/internal/mqcmd"

	"github.com/spf13/cobra"
)

var (
	mqDir        string
	mqConfig     string
	mqEventType  string
	mqTopic      string
	mqPayload    string
	mqSource     string
	mqTraceID    string
)

var mqCmd = &cobra.Command{
	Use:   "mq",
	Short: "Message queue utilities (NSQ)",
}

var mqPublishCmd = &cobra.Command{
	Use:   "publish",
	Short: "Publish an event envelope to NSQ using project config",
	RunE: func(cmd *cobra.Command, args []string) error {
		abs, err := filepath.Abs(mqDir)
		if err != nil {
			abs = mqDir
		}
		return mqcmd.Publish(mqcmd.PublishOptions{
			Dir:        abs,
			ConfigPath: resolvedConfigPath(abs, mqConfig),
			EventType:  mqEventType,
			Topic:      mqTopic,
			Payload:    mqPayload,
			Source:     mqSource,
			TraceID:    mqTraceID,
		})
	},
}

func init() {
	rootCmd.AddCommand(mqCmd)
	mqCmd.PersistentFlags().StringVar(&mqDir, "dir", ".", "Project root directory")
	mqCmd.PersistentFlags().StringVarP(&mqConfig, "config", "f", "", "Config file path (default: configs/config.yaml or GOEASY_CONFIG)")

	mqPublishCmd.Flags().StringVar(&mqEventType, "event-type", "", "Event type (also used as NSQ topic when --topic omitted)")
	mqPublishCmd.Flags().StringVar(&mqTopic, "topic", "", "NSQ topic override")
	mqPublishCmd.Flags().StringVar(&mqPayload, "payload", "{}", "JSON payload object")
	mqPublishCmd.Flags().StringVar(&mqSource, "source", "goeasy-cli", "source_service field")
	mqPublishCmd.Flags().StringVar(&mqTraceID, "trace-id", "", "trace_id (auto-generated when empty)")
	_ = mqPublishCmd.MarkFlagRequired("event-type")

	mqCmd.AddCommand(mqPublishCmd)
}
