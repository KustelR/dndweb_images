package main

import (
	"fmt"
	"hash/fnv"
	"io"
	"net/http"
	"os"
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
		if contentType != "image/jpeg" {
			w.WriteHeader(http.StatusBadRequest)
		}
		fmt.Fprint(w, createAsset(r.Body))
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

func createAsset(data io.ReadCloser) string {
	byteData, err := io.ReadAll(data)
	if err != nil {
		fmt.Printf("unable to read asset: %s\n", err)
	}
	hashed := fnv.New32a()
	filename := strconv.FormatInt(int64(fnv.New32a().Sum32()), 16)
	hashed.Write(byteData)
	err = os.WriteFile(fmt.Sprintf("./static/%v", filename), byteData, 0777)
	if err != nil {
		fmt.Printf("unable to create asset: %s\n", err)
	}
	return filename
}
