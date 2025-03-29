package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/lovoo/goka"
)

var (
	Group       goka.Group        = "metrics-group"
	lastMetrics map[string]Metric = make(map[string]Metric)
)

type Metric struct {
	Type        string
	Name        string
	Description string
	Value       float64
}

type MetricCodec struct{}

func (uc *MetricCodec) Encode(value any) ([]byte, error) {
	if _, isConsumer := value.(*map[string]Metric); !isConsumer {
		return nil, fmt.Errorf("expected type is *Metric, received %T", value)
	}
	return json.Marshal(value)
}

func (uc *MetricCodec) Decode(data []byte) (any, error) {
	var (
		c   map[string]Metric
		err error
	)
	err = json.Unmarshal(data, &c)
	if err != nil {
		return nil, fmt.Errorf("deserialization error: %v", err)
	}
	return &c, nil
}

func process(ctx goka.Context, msg any) {
	var metrics *map[string]Metric
	var ok bool

	if metrics, ok = msg.(*map[string]Metric); !ok || metrics == nil {
		return
	}

	var copied = CopyMetrics()

	for key, metric := range *metrics {
		copied[key] = metric
	}

	lastMetrics = copied

	log.Printf("[proc] msg: %+v\n", lastMetrics)
}

func CopyMetrics() map[string]Metric {
	var copied = make(map[string]Metric, len(lastMetrics))
	for key, metric := range lastMetrics {
		copied[key] = metric
	}
	return copied
}

func RunConsumerProcessor(brokers []string, inputStream goka.Stream) {
	g := goka.DefineGroup(Group,
		goka.Input(inputStream, new(MetricCodec), process),
		goka.Persist(new(MetricCodec)),
	)

	for true {
		p, err := goka.NewProcessor(brokers, g)
		if err != nil {
			log.Printf("Error creating the processor: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}
		err = p.Run(context.Background())
		if err != nil {
			log.Printf("Error running the goka context: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}
	}
}
