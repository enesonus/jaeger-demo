package main

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)

// album represents data about a record album.
type album struct {
	ID     string  `json:"id"`
	Title  string  `json:"title"`
	Artist string  `json:"artist"`
	Price  float64 `json:"price"`
}

func main() {
    router := gin.Default()
    router.GET("/albums/:id", getAlbumByID)

    router.Run("localhost:8080")
}

// getAlbumByID locates the album whose ID value matches the id
// parameter sent by the client, then returns that album as a response.
func getAlbumByID(c *gin.Context) {
    id := c.Param("id")

	result := album{
		ID:     id,
		Title:  getTitleById(id, c),
		Artist: getArtistById(id, c),
		Price:  getPriceById(id, c),
	}

    c.IndentedJSON(http.StatusOK, result)
    return
}

func getPriceById(id string, c *gin.Context) float64 {
    var priceStruct struct {Price float64 `json:"price"`}

    if err := fetchJSON("http://localhost:8081/album_price/"+id, &priceStruct); err != nil {
        c.IndentedJSON(http.StatusNotFound, gin.H{"Error fetching price:": err})
        return 0
    }
    return priceStruct.Price
}

func getTitleById(id string, c *gin.Context) string {
    var titleStruct struct {Title string `json:"title"`}

    if err := fetchJSON("http://localhost:8082/album_title/"+id, &titleStruct); err != nil {
        c.IndentedJSON(http.StatusNotFound, gin.H{"Error fetching title:": err})
        return ""
    }
    return titleStruct.Title
}

func getArtistById(id string, c *gin.Context) string {
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

