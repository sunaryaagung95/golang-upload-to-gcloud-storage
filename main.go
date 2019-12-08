package main

import (
	"context"
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)



var gcsBucket string



func main() {
	loadENV()
	gcsBucket = os.Getenv("BUCKET_NAME")	
	router := mux.NewRouter()

	router.HandleFunc("/upload", getFile).Methods("POST")

	log.Fatal(http.ListenAndServe(":8080", router))
}

func loadENV() {
	var err error
	err = godotenv.Load()
	if err != nil {
		panic(err)
	}
	
}

func getFile(w http.ResponseWriter, r *http.Request) {
	mpf, hdr, err := r.FormFile("file")
	if err != nil {
		panic(err)
	}
	defer mpf.Close()

	fname, err := uploadFile(r, mpf, hdr)
	if err != nil {
		panic(err)
	}
	fileURL := "https://storage.cloud.google.com/" + gcsBucket + `/` + fname

	json.NewEncoder(w).Encode(fileURL)
}

func uploadFile(r *http.Request, mpf multipart.File, hdr *multipart.FileHeader) (string, error) {
	ext, err := fileFilter(r, hdr)
	if err != nil {
		return "", err
	}
	name := getSha(mpf) + `.` + ext
	mpf.Seek(0, 0)

	ctx := context.Background()
	return name, putFile(ctx, name, mpf)

}

func fileFilter(r *http.Request, hdr *multipart.FileHeader) (string, error) {
	ext := hdr.Filename[strings.LastIndex(hdr.Filename, ".")+1:]

	switch strings.ToLower(ext) {
	case "jpg", "jpeg", "png":
		return ext, nil
	}
	return ext, errors.New("Not supported type")
}

func getSha(src multipart.File) string {
	h := sha1.New()
	io.Copy(h, src)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func putFile(ctx context.Context, name string, rdr io.Reader) error {

	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}
	defer client.Close()

	writer := client.Bucket(gcsBucket).Object(name).NewWriter(ctx)

	io.Copy(writer, rdr)
	return writer.Close()
}
