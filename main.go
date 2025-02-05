package main

import (
	"fmt"
	"hash/fnv"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
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

func main() {
	NewServer(80, HandleRequest)
}

func createAsset(data io.ReadCloser) (string, error) {
	byteData, err := io.ReadAll(data)
	if err != nil {
		fmt.Printf("unable to read asset: %s\n", err)
	}
	isImage, _ := checkImageType(http.DetectContentType(byteData))
	if !isImage {
		return "", fmt.Errorf("provided file is not an image")
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
