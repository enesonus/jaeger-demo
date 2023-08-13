package main

import (
	"context"
	"net/http"

	"github.com/enesonus/jaeger-demo/lib/tracing"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// album represents data about a record album.
type album struct {
	ID     string  `json:"id"`
	Title  string  `json:"title"`
	Artist string  `json:"artist"`
	Price  float64 `json:"price"`
}


// albums slice to seed record album data.
var albums = []album{
    {ID: "1", Title: "Blue Train", Artist: "John Coltrane", Price: 56.99},
    {ID: "2", Title: "Jeru", Artist: "Gerry Mulligan", Price: 17.99},
    {ID: "3", Title: "Sarah Vaughan and Clifford Brown", Artist: "Sarah Vaughan", Price: 39.99},
}

var tracer trace.Tracer

const thisServiceName = "price-service"

func main() {
	ctx := context.Background()
    tracer = tracing.Init(ctx, thisServiceName)


    router := gin.New()
	router.Use(otelgin.Middleware(thisServiceName))
    router.GET("/album_price/:id", getAlbumPriceByID)

    router.Run("localhost:8081")
}

// getAlbumByID locates the album whose ID value matches the id
// parameter sent by the client, then returns that album as a response.
func getAlbumPriceByID(c *gin.Context) {
	p := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{}, propagation.Baggage{})
		
	ctx := p.Extract(c.Request.Context(), propagation.HeaderCarrier(c.Request.Header))

	_, span := tracer.Start(ctx, thisServiceName)
	defer span.End()

	c.Request.Header.Set("Content-Type", "application/json")

    id := c.Param("id")

    // Loop over the list of albums, looking for
    // an album whose ID value matches the parameter.
    for _, a := range albums {
        if a.ID == id {
            c.IndentedJSON(http.StatusOK, gin.H{"price": a.Price})
            return
        }
    }
    c.IndentedJSON(http.StatusNotFound, gin.H{"message": "Price not found"})
}