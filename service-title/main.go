package main

import (
	"context"
	"github.com/enesonus/jaeger-demo/internal/tracing"
	"net/http"

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

const thisServiceName = "title-service"

func main() {
	ctx := context.Background()
	tracer = tracing.Init(ctx, thisServiceName)

	router := gin.New()
	router.Use(otelgin.Middleware(thisServiceName))
	router.GET("/album_title/:id", getAlbumTitleByID)

	router.Run("localhost:8082")
}

// getAlbumByID locates the album whose ID value matches the id
// parameter sent by the client, then returns that album as a response.
func getAlbumTitleByID(c *gin.Context) {

	p := propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{})

	ctx := p.Extract(c.Request.Context(), propagation.HeaderCarrier(c.Request.Header))

	_, span := tracer.Start(ctx, "service-title")
	defer span.End()

	c.Request.Header.Set("Content-Type", "application/json")

	id := c.Param("id")

	// Loop over the list of albums, looking for
	// an album whose ID value matches the parameter.
	for _, a := range albums {
		if a.ID == id {
			c.IndentedJSON(http.StatusOK, gin.H{"title": a.Title})
			return
		}
	}
	c.IndentedJSON(http.StatusNotFound, gin.H{"message": "Title not found"})
}
