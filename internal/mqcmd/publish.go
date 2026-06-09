package mqcmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/txbao/goeasy/config"
	"github.com/txbao/goeasy/mq"
)

// PublishOptions mq publish 命令参数。
type PublishOptions struct {
	Dir        string
	ConfigPath string
	EventType  string
	Topic      string
	Payload    string
	Source     string
	TraceID    string
}

// Publish 读取项目配置并向 NSQ 发布标准信封消息。
func Publish(opts PublishOptions) error {
	cfgPath := opts.ConfigPath
	if !filepath.IsAbs(cfgPath) {
		cfgPath = filepath.Join(opts.Dir, cfgPath)
	}
	cfg := config.MustLoad(cfgPath)
	if !cfg.MQ.Enabled {
		return fmt.Errorf("mq.enabled is false in %s", cfgPath)
	}
	if opts.EventType == "" {
		return fmt.Errorf("--event-type is required")
	}
	topic := opts.Topic
	if topic == "" {
		topic = opts.EventType
	}
	source := opts.Source
	if source == "" {
		source = "goeasy-cli"
	}
	var payload json.RawMessage
	if opts.Payload == "" {
		payload = mq.EmptyPayload()
	} else {
		payload = json.RawMessage(opts.Payload)
		if !json.Valid(payload) {
			return fmt.Errorf("invalid --payload JSON")
		}
	}
	client, err := mq.Open(cfg.MQ)
	if err != nil {
		return err
	}
	defer client.Close()

	env, err := mq.NewEnvelope(opts.EventType, source, opts.TraceID, payload)
	if err != nil {
		return err
	}
	body, err := env.Marshal()
	if err != nil {
		return err
	}
	if err := client.Publish(context.Background(), topic, body); err != nil {
		return err
	}
	fmt.Fprintf(os.Stdout, "published topic=%s event_id=%s event_type=%s payload_bytes=%d\n",
		topic, env.EventID, env.EventType, env.PayloadBytes())
	return nil
}
