package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"restful-api/lib/tracing"

	"net/http"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	oteltrace "go.opentelemetry.io/otel/trace"
)

const name = "jaeger-tracer"

// album represents data about a record album.
type album struct {
	ID     string  `json:"id"`
	Title  string  `json:"title"`
	Artist string  `json:"artist"`
	Price  float64 `json:"price"`
}


// func NewTracerProvider(ctx context.Context) *sdktrace.TracerProvider {

// 	exporter, err := jaeger.New(jaeger.WithCollectorEndpoint(
// 		jaeger.WithEndpoint("http://localhost:14268/api/traces"),
// 	),
// 	)
// 	if err != nil {
// 		panic(err)
// 	}

// 	// Ensure default SDK resources and the required service name are set.
// 	r, err := resource.Merge(
// 		resource.Default(),
// 		resource.NewWithAttributes(
// 			semconv.SchemaURL,
// 			semconv.ServiceName("jaeger-demo-eonus"),
// 			semconv.ServiceVersionKey.String("1.0.0"),
// 			semconv.ServiceInstanceIDKey.String("abcdef12345"),
// 		),
// 	)

// 	if err != nil {
// 		panic(err)
// 	}

// 	return sdktrace.NewTracerProvider(
// 		sdktrace.WithSampler(sdktrace.AlwaysSample()),
// 		sdktrace.WithBatcher(exporter),
// 		sdktrace.WithResource(r),
// 	)
// }

const thisServiceName = "jaeger-demo-eonus-main"

var tracer trace.Tracer

func main() {

	ctx := context.Background()

	tracer = tracing.Init(ctx, thisServiceName) 

	http.HandleFunc("/album", getAlbumByID)
	log.Printf("Listening on localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
	
}

// getAlbumByID locates the album whose ID value matches the id
// parameter sent by the client, then returns that album as a response.
func getAlbumByID(w http.ResponseWriter, r *http.Request) {

	ctx, span := tracer.Start(r.Context(), "thisServiceName")
	defer span.End()

	id := r.URL.Query().Get("id")

	pricechan := make(chan float64)
	titlechan := make(chan string)
	artistchan := make(chan string)

	go func() {
		pricechan <- getPriceById(ctx, id)
	}()

	go func() {
		titlechan <- getTitleById(ctx, id)
	}()

	go func() {
		artistchan <- getArtistById(ctx, id)
	}()

	price := <-pricechan

	title := <-titlechan

	artist := <-artistchan

	result := album{
		ID:     id,
		Title:  title,
		Artist: artist,
		Price:  price,
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	jsonResp, err := json.Marshal(result)
	if err != nil {
		log.Fatalf("Error happened in JSON marshal. Err: %s", err)
	}
	w.Write(jsonResp)
	return
}

func getPriceById(ctx context.Context, id string) float64 {
	ctx, span := tracer.Start(ctx, "getPriceById",
		oteltrace.WithAttributes(attribute.String("id", id)))
	defer span.End()

	var priceStruct struct {
		Price float64 `json:"price"`
	}

	if err := fetchJSON("http://localhost:8081/album_price/"+id, &priceStruct, ctx); err != nil {
		return 0
	}
	return priceStruct.Price
}

func getTitleById(ctx context.Context, id string) string {
	ctx, span := tracer.Start(ctx, "getTitleById",
		oteltrace.WithAttributes(attribute.String("id", id)))

	defer span.End()

	var titleStruct struct {
		Title string `json:"title"`
	}

	if err := fetchJSON("http://localhost:8082/album_title/"+id, &titleStruct, ctx); err != nil {
		return ""
	}
	return titleStruct.Title
}

func getArtistById(ctx context.Context, id string) string {

	ctx, span := tracer.Start(ctx, "getArtistById",
		oteltrace.WithAttributes(attribute.String("id", id)))

	defer span.End()

	var artistStruct struct {
		Artist string `json:"artist"`
	}
	url := "http://localhost:8083/album_artist?id="+id	

	if err := fetchJSON(url, &artistStruct, ctx); err != nil {
		return err.Error()
	}
	return artistStruct.Artist
}

func fetchJSON(url string, target interface{}, ctx context.Context) error {
	

	ctx, span := tracer.Start(ctx, "fetchJSON")
	defer span.End()


	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)

	propagator := propagation.TraceContext{}
	propagator.Inject(ctx, propagation.HeaderCarrier(req.Header))

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	err = json.Unmarshal(respBody, target)
	return nil
}
