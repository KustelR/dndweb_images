package main

import (
	"fmt"
	"hash/fnv"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"

	"github.com/joho/godotenv"
)

type HttpServer struct {
	Handler func(http.ResponseWriter, *http.Request)
}

func HandleRequest(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		contentType := r.Header.Get("Content-Type")

		fmt.Printf("POST request received  Content-Type: %s\n", contentType)
		isImage, _ := checkImageType(contentType)
		filename, err := createAsset(r.Body)
		if !isImage || err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "Invalid content type. Only images are allowed.")
			return
		}
		fmt.Fprint(w, filename)
		w.WriteHeader(http.StatusCreated)
	}
}

func NewServer(port int, handler func(http.ResponseWriter, *http.Request)) {

	http.HandleFunc("/", HandleRequest)
	fs := http.FileServer(http.Dir("static/"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	fmt.Printf("Server listening at %v\n", port)
	err := http.ListenAndServe(fmt.Sprintf(":%v", port), nil)
	if err != nil {
		panic(err)
	}

}

func init() {
	// loads values from .env into the system
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
}

func main() {
	var port int
	imagesApiPort, iapExists := os.LookupEnv("IMAGES_API_PORT")
	if iapExists {
		parsedPort, _ := strconv.Atoi(imagesApiPort)
		port = parsedPort
	}
	if port == 0 {
		panic("port is not provided: set env IMAGES_API_PORT")
	}

	NewServer(port, HandleRequest)
}

func createAsset(data io.ReadCloser) (string, error) {
	byteData, err := io.ReadAll(data)
	if err != nil {
		fmt.Printf("unable to read asset: %s\n", err)
	}
	contentType := http.DetectContentType(byteData)
	isImage, _ := checkImageType(contentType)
	if !isImage {
		return "", fmt.Errorf("provided file is not an image, it is: %s", contentType)
	}
	hashed := fnv.New32a()
	hashed.Write(byteData)
	filename := strconv.FormatInt(int64(hashed.Sum32()), 16)
	err = os.WriteFile(fmt.Sprintf("./static/%v", filename), byteData, 0777)
	if err != nil {
		fmt.Printf("unable to create asset: %s\n", err)
	}
	return filename, nil
}
func checkImageType(contentType string) (bool, error) {
	return regexp.MatchString("^image", contentType)
}

func cors(handler func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if r.Method == "OPTIONS" {
			return
		}

		handler(w, r)
	}
}
