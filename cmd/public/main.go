package main

import (
	"context"
	"encoding/json"
	"fmt"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	oteltrace "go.opentelemetry.io/otel/trace"
	"golang.org/x/sync/errgroup"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/enesonus/jaeger-demo/internal/tracing"
	"github.com/enesonus/jaeger-demo/pkg/models"
)

const (
	ServiceName = "public"

	EnvVariablePort             = "PORT"
	EnvVariableServiceArtistURL = "SERVICE_ARTIST_URL"
	EnvVariableServicePriceURL  = "SERVICE_PRICE_URL"
	EnvVariableServiceTitleURL  = "SERVICE_TITLE_URL"

	DefaultPort = "8080"
)

var tracer trace.Tracer

func main() {
	ctx := context.Background()
	tracer = tracing.Init(ctx, ServiceName)

	svcURLArtist, ok := os.LookupEnv(EnvVariableServiceArtistURL)
	if !ok {
		log.Fatalf("%s is required", EnvVariableServiceArtistURL)
	}
	svcURLPrice, ok := os.LookupEnv(EnvVariableServicePriceURL)
	if !ok {
		log.Fatalf("%s is required", EnvVariableServicePriceURL)
	}
	svcURLTitle, ok := os.LookupEnv(EnvVariableServiceTitleURL)
	if !ok {
		log.Fatalf("%s is required", EnvVariableServiceTitleURL)
	}

	server := http.NewServeMux()
	server.HandleFunc("/album", func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tracer.Start(r.Context(), ServiceName)
		defer span.End()
		w.Header().Set("Content-Type", "application/json")
		encoder := json.NewEncoder(w)

		id := r.URL.Query().Get("id")
		var price float64
		var title string
		errGroup := new(errgroup.Group)
		errGroup.Go(func() (err error) {
			price, err = getPriceById(ctx, svcURLPrice, id)
			return
		})
		errGroup.Go(func() (err error) {
			title, err = getTitleById(ctx, svcURLTitle, id)
			return
		})
		if err := errGroup.Wait(); err != nil {
			if err := encoder.Encode(map[string]string{"error": err.Error()}); err != nil {
				fmt.Printf("cannot write back error as response %s", err)
			}
			w.WriteHeader(http.StatusInternalServerError)
		}
		// This is left to be serial execution to show a different trace pattern.
		artist, err := getArtistById(ctx, svcURLArtist, id)
		if err != nil {
			if err := encoder.Encode(map[string]string{"error": err.Error()}); err != nil {
				fmt.Printf("cannot write back error as response %s", err)
			}
			w.WriteHeader(http.StatusInternalServerError)
		}
		result := models.Album{
			ID:     id,
			Title:  title,
			Artist: artist,
			Price:  price,
		}
		if err := encoder.Encode(result); err != nil {
			if err := encoder.Encode(map[string]string{"error": err.Error()}); err != nil {
				fmt.Printf("cannot write back error as response %s", err)
			}
			w.WriteHeader(http.StatusInternalServerError)
		}
		w.WriteHeader(http.StatusOK)
	})
	port, ok := os.LookupEnv(EnvVariablePort)
	if !ok {
		port = DefaultPort
	}
	log.Printf("Listening on %s", port)
	log.Fatal(http.ListenAndServe(":"+port, server))

}

func getPriceById(ctx context.Context, svcURL, id string) (float64, error) {
	ctx, span := tracer.Start(ctx, "getPriceById",
		oteltrace.WithAttributes(attribute.String("id", id)))
	defer span.End()

	artistResp, err := fetchJSON(ctx, fmt.Sprintf("%s/album_price?id=%v", svcURL, id))
	if err != nil {
		return 0, fmt.Errorf("failed to fetch price: %w", err)
	}
	f, err := strconv.ParseFloat(artistResp["price"], 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse price: %w", err)
	}
	return f, nil
}

func getTitleById(ctx context.Context, svcURL, id string) (string, error) {
	ctx, span := tracer.Start(ctx, "getTitleById",
		oteltrace.WithAttributes(attribute.String("id", id)))
	defer span.End()

	artistResp, err := fetchJSON(ctx, fmt.Sprintf("%s/album_title?id=%v", svcURL, id))
	if err != nil {
		return "", fmt.Errorf("failed to fetch title: %w", err)
	}
	return artistResp["title"], nil
}

func getArtistById(ctx context.Context, svcURL, id string) (string, error) {
	ctx, span := tracer.Start(ctx, "getArtistById",
		oteltrace.WithAttributes(attribute.String("id", id)))
	defer span.End()

	artistResp, err := fetchJSON(ctx, fmt.Sprintf("%s/album_artist?id=%v", svcURL, id))
	if err != nil {
		return "", fmt.Errorf("failed to fetch artist: %w", err)
	}
	return artistResp["artist"], nil
}

func fetchJSON(ctx context.Context, url string) (map[string]string, error) {
	ctx, span := tracer.Start(ctx, "fetchJSON")
	defer span.End()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	propagator := propagation.TraceContext{}
	propagator.Inject(ctx, propagation.HeaderCarrier(req.Header))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch url: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		errBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read error body: %w", err)
		}
		res := map[string]string{}
		if err = json.Unmarshal(errBody, &res); err != nil {
			return nil, fmt.Errorf("failed to unmarshal error response body: %w", err)
		}
		return nil, fmt.Errorf("failed to fetch url: %s", res["error"])
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	res := map[string]string{}
	if err = json.Unmarshal(respBody, &res); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %w", err)
	}
	return res, nil
}
