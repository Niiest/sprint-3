package main

import (
	"example/connector/consumer"
	"fmt"
	"log"
	"net/http"
	"os"
	"text/template"

	"github.com/IBM/sarama"
	"github.com/go-chi/chi/v5"
	"github.com/lovoo/goka"
)

const metricsTemplate = `{{ range $key, $value := . }}# HELP {{$value.Description}}
# TYPE {{$value.Name}} {{$value.Type}}
{{ $key }} {{ $value.Value }}
{{ end }}`

var (
	brokers                   = []string{os.Getenv("KAFKA_URL")}
	MetricsStream goka.Stream = "metrics"
)

func main() {
	log.Printf("Brokers %v", brokers)

	config := goka.DefaultConfig()
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Consumer.Offsets.AutoCommit.Enable = false
	goka.ReplaceGlobalConfig(config)

	go consumer.RunConsumerProcessor(brokers, MetricsStream)

	// Создание и парсинг шаблона
	tmpl, err := template.New("metrics").Parse(metricsTemplate)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка при парсинге шаблона: %v\n", err)
		os.Exit(1)
	}

	r := chi.NewRouter()
	r.Get("/", func(rw http.ResponseWriter, r *http.Request) {
		rw.Write([]byte("chi"))
	})
	r.Get("/metrics", func(rw http.ResponseWriter, r *http.Request) {
		println("Received metrics request")

		var metrics = consumer.CopyMetrics()

		// Выполнение шаблона и отдача результата
		if err := tmpl.Execute(rw, metrics); err != nil {
			fmt.Fprintf(os.Stderr, "Ошибка при выполнении шаблона: %v\n", err)
			os.Exit(1)
		}
	})

	http.ListenAndServe(":8084", r)
}
