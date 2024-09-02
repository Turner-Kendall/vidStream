package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tr1pwyr/go-video-streaming/controllers"
)

type Category string

const (
	Funny    Category = "Funny"
	Podcast  Category = "Podcast"
	Rant     Category = "Rants"
	Viral    Category = "Viral"
	Gaming   Category = "Gaming"
	Music    Category = "Music"
	Sports   Category = "Sports"
	Trending Category = "Trending"
	News     Category = "News"
	Tech     Category = "Tech"
	Travel   Category = "Travel"
)

type PageData struct {
	Categories []Category
}

type Video struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Creator     string   `json:"creator"`
	Filename    string   `json:"filename"`
	Timestamp   int64    `json:"timestamp"`
	Thumbnail   string   `json:"thumbnail"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	Category    Category `json:"category"`
}

var videos []Video

func main() {

	var err error
	videos, err = loadJson()

	if err != nil {
		fmt.Println(err)
	}

	router := gin.Default()
	router.LoadHTMLGlob("templates/*")
	router.MaxMultipartMemory = 8 << 20 // 8 MiB

	router.Static("/videoUpload", "./public/upload")

	router.Static("/assets", "./assets")
	router.GET("/", renderVideos)

	router.GET("/stream/:filename", streamFile)
	router.GET("/videos", getVideos)
	router.GET("/video/:id", getVideoById)

	router.GET("/upload-form", uploadVideoForm)
	router.POST("/upload", uploadVideo)

	router.Use(gin.Recovery())
	router.Run(":8080")

}

func streamFile(c *gin.Context) {
	filename := c.Param("filename")
	file, err := os.Open("videos/" + filename)
	if err != nil {
		c.String(http.StatusNotFound, "Video not found.")
		return
	}
	defer file.Close()
	c.Header("Content-Type", "video/mp4")
	buffer := make([]byte, 64*1024) // 64KB buffer size
	io.CopyBuffer(c.Writer, file, buffer)
}

func getVideos(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, videos)
}

func getVideoById(c *gin.Context) {
	id := c.Param("id")
	for _, a := range videos {
		if a.ID == id {
			c.IndentedJSON(http.StatusOK, a)
			return
		}
	}
	c.IndentedJSON(http.StatusNotFound, gin.H{"message": "video not found"})
}

func renderVideos(c *gin.Context) {
	c.HTML(http.StatusOK, "videos.tmpl", gin.H{
		"title":    "Vstrem",
		"subTitle": "Simple, fast, video streaming.",
		"videos":   videos,
	})
}

func uploadVideoForm(c *gin.Context) {

	categories := []Category{Funny, Rant, Podcast /* Add more categories here */}

	c.HTML(http.StatusOK, "upload.tmpl", gin.H{
		"title": "Vstrem Upload Video.",
		"data":  categories,
		// "videos":   videos,
	})
}

func uploadVideo(c *gin.Context) {
	name := c.PostForm("name")
	email := c.PostForm("email")
	timestamp := time.Now().Unix()
	description := c.PostForm("description")
	// t := time.Unix(timestamp, 0)

	// tags := []string{"video", "test", "tag"}

	//OR
	// 	var tags []string
	// // Add tags based on user input or other sources
	// tags = append(tags, "tag1")
	// tags = append(tags, "tag2")
	// tags = append(tags, "tag3")

	// id := generateId()
	id := controllers.GenerateId()

	file, err := c.FormFile("file")

	if err != nil {
		c.String(http.StatusBadRequest, "get form err: %s", err.Error())
		return
	}

	uploadDir := "staging"
	destination := filepath.Join(uploadDir, file.Filename)

	if err := c.SaveUploadedFile(file, destination); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	newVideo := Video{
		ID:          string(id),
		Title:       name,
		Creator:     email,
		Filename:    file.Filename,
		Timestamp:   timestamp,
		Thumbnail:   "default.webp",
		Description: description,
		Tags:        []string{"tag1", "tag2", "tag3"}, // Example tags
		Category:    Funny,                            // Example category
	}

	videos = append(videos, newVideo)

	if err := saveJson(videos); err != nil {
		fmt.Printf("%s", err)
	}

	c.String(http.StatusOK, "File %s uploaded successfully named=%s with an id=%s.", file.Filename, name, id)
}

func loadJson() ([]Video, error) {

	var v []Video

	jsonFile, err := os.Open("videos.json")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Successfully Opened users.json")
	defer jsonFile.Close()

	byteValue, _ := io.ReadAll(jsonFile)
	json.Unmarshal(byteValue, &v)
	return v, nil
}

func saveJson(v []Video) error {
	// Marshal the list of videos into JSON format
	jsonData, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("error marshaling JSON: %v", err)
	}

	// Open the JSON file for writing
	jsonFile, err := os.Create("videos.json")
	if err != nil {
		return fmt.Errorf("error creating JSON file: %v", err)
	}
	defer jsonFile.Close()

	// Write the JSON data to the file
	_, err = jsonFile.Write(jsonData)
	if err != nil {
		return fmt.Errorf("error writing JSON data to file: %v", err)
	}

	fmt.Println("JSON data has been written to file successfully")
	return nil
}
