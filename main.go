package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

const name = "jaeger-tracer"
// album represents data about a record album.
type album struct {
	ID     string  `json:"id"`
	Title  string  `json:"title"`
	Artist string  `json:"artist"`
	Price  float64 `json:"price"`
}

var tracer = otel.Tracer(name)

func main() {
	os.Setenv("OTEL_SERVICE_NAME", "jaeger-demo")
	tp, err := initTracer()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}()

    router := gin.New()
	router.Use(otelgin.Middleware("jaeger-demo"))
    router.GET("/albums/:id", getAlbumByID)

    router.Run("localhost:8080")
}

func initTracer() (*sdktrace.TracerProvider, error) {
	
    // Create the Jaeger exporter
    exporter, err := jaeger.New(
        jaeger.WithCollectorEndpoint(jaeger.WithEndpoint("http://localhost:14268/api/traces")),
    )
    if err != nil {
        return nil, err
    }

    // Create a new TracerProvider
    tp := sdktrace.NewTracerProvider(
        sdktrace.WithSampler(sdktrace.AlwaysSample()),
        sdktrace.WithBatcher(exporter),
    )

    // Set the global TracerProvider and the propagator
    otel.SetTracerProvider(tp)
    otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

    return tp, nil
}


// getAlbumByID locates the album whose ID value matches the id
// parameter sent by the client, then returns that album as a response.
func getAlbumByID(c *gin.Context) {

    id := c.Param("id")

	price := getPriceById(id, c)
	
	title := getTitleById(id, c)
	
	artist := getArtistById(id, c)

	result := album{
		ID:     id,
		Title:  title,
		Artist: artist,
		Price:  price,
	}

    c.IndentedJSON(http.StatusOK, result)
    return
}



func getPriceById(id string, c *gin.Context) float64 {
	_, span_price := tracer.Start(c.Request.Context(), "get-price-span")
	defer span_price.End()
	
    var priceStruct struct {Price float64 `json:"price"`}

    if err := fetchJSON("http://localhost:8081/album_price/"+id, &priceStruct); err != nil {
        c.IndentedJSON(http.StatusNotFound, gin.H{"Error fetching price:": err})
        return 0
    }
    return priceStruct.Price
}

func getTitleById(id string, c *gin.Context) string {
	_, span_title := tracer.Start(c.Request.Context(), "get-title-span")
	defer span_title.End()

    var titleStruct struct {Title string `json:"title"`}

    if err := fetchJSON("http://localhost:8082/album_title/"+id, &titleStruct); err != nil {
        c.IndentedJSON(http.StatusNotFound, gin.H{"Error fetching title:": err})
        return ""
    }
    return titleStruct.Title
}

func getArtistById(id string, c *gin.Context) string {
	_, span_artist := tracer.Start(c.Request.Context(), "get-artist-span")
	defer span_artist.End()

    var artistStruct struct {Artist string `json:"artist"`}

	if err := fetchJSON("http://localhost:8083/album_artist/"+id, &artistStruct); err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"Error fetching artist:": err})
		return ""
	}
    return artistStruct.Artist
}

func fetchJSON(url string, target interface{}) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return json.NewDecoder(resp.Body).Decode(target)
}

