package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"go.opentelemetry.io/otel/propagation"

	"github.com/enesonus/jaeger-demo/internal/tracing"
	"github.com/enesonus/jaeger-demo/pkg/models"
)

const (
	ServiceName = "service-price"
	SpanNameFmt = ServiceName + "/%s"

	EnvVariablePort = "PORT"

	DefaultPort = "8080"
)

func main() {
	ctx := context.Background()
	tracer := tracing.Init(ctx, ServiceName)
	server := http.NewServeMux()
	server.HandleFunc("/album_price", func(w http.ResponseWriter, r *http.Request) {
		p := propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{})
		ctx := p.Extract(r.Context(), propagation.HeaderCarrier(r.Header))
		_, span := tracer.Start(ctx, fmt.Sprintf(SpanNameFmt, "album_price"))
		defer span.End()

		w.Header().Set("Content-Type", "application/json")
		encoder := json.NewEncoder(w)

		id := r.URL.Query().Get("id")
		for _, album := range models.AlbumsSeed {
			if album.ID == id {
				if err := encoder.Encode(map[string]string{"price": fmt.Sprintf("%f", album.Price)}); err != nil {
					if err := encoder.Encode(map[string]string{"error": err.Error()}); err != nil {
						fmt.Printf("cannot write back error as response %s", err)
					}
				}
				return
			}
		}
		w.WriteHeader(http.StatusNotFound)
	})
	port, ok := os.LookupEnv(EnvVariablePort)
	if !ok {
		port = DefaultPort
	}
	log.Printf("Listening on %s", port)
	log.Fatal(http.ListenAndServe(":"+port, server))
}
