package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"restful-api/lib/tracing"

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
const thisServiceName = "artist-service"
func main() {
	
	ctx := context.Background()
	tracer = tracing.Init(ctx, thisServiceName) 
	http.HandleFunc("/album_artist", getAlbumArtistByID)

	log.Printf("Listening on localhost:8083")
	log.Fatal(http.ListenAndServe(":8083", nil))
}

// getAlbumArtistByID locates the artist whose ID value matches the id
// parameter sent by the client, then returns that artist as a response.
func getAlbumArtistByID(w http.ResponseWriter, r *http.Request) {

	p := propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{})
		
	ctx := p.Extract(r.Context(), propagation.HeaderCarrier(r.Header))

	_, span := tracer.Start(ctx, "service-artist")
	defer span.End()

	// w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")

	id := r.URL.Query().Get("id")

	for _, album := range albums {
		if album.ID == id {
			fmt.Println("Artist: ", album.Artist)

			w.Header().Set("Content-Type", "application/json")

			jsonResp, err := json.Marshal(album)
			if err != nil {
				log.Fatalf("Error happened in JSON marshal. Err: %s", err)
			}

			w.Write(jsonResp)

			return
		}
	}
	
}
