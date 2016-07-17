package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	currentDir, _ := os.Getwd()

	fmt.Println("Listening...")
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		path := strings.TrimPrefix(request.URL.Path, "/")
		fmt.Println(path)

		requestTime := time.Now().Format("2006-01-02T15_04_05.999")
		outDir := filepath.Join(currentDir, path, requestTime)
		os.MkdirAll(outDir, os.ModeDir)

		contentType := request.Header.Get("Content-Type")

		if strings.Index(contentType, "multipart/form-data") == 0 {
			handleMultipartFormUpload(outDir, request)
		} else {
			handlePostBodyUpload(outDir, request)
		}

	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleMultipartFormUpload(outDir string, request *http.Request) {
	if err := request.ParseMultipartForm(10000000); err != nil {
		fmt.Println("Error while parsing multipart form: " + err.Error())
		return
	}

	form := request.MultipartForm

	for _, file := range form.File["data-file"] {
		f, err := file.Open()
		defer f.Close()

		if err != nil {
			fmt.Println("Error while opening file " + err.Error())
			continue
		}

		outPath := filepath.Join(outDir, file.Filename)

		outFile, err := os.Create(outPath)
		defer outFile.Close()

		if err != nil {
			fmt.Println("Error while creating output file " + err.Error())
			continue
		}

		if _, err := io.Copy(outFile, f); err != nil {
			fmt.Println("Error while writing to output file " + err.Error())
			continue
		}

		saveFormData(outDir, request)

		fmt.Println("File downloaded! " + file.Filename)
	}
}

func saveFormData(outDir string, request *http.Request) {
	request.ParseForm()
	for k, value := range request.Form {
		outDataPath := filepath.Join(outDir, k+".txt")
		ioutil.WriteFile(outDataPath, []byte(strings.Join(value, ",")), 0644)
	}
}

func handlePostBodyUpload(outDir string, request *http.Request) {
	fileName := getFileName(request)
	outFile, err := os.Create(filepath.Join(outDir, fileName))
	defer outFile.Close()
	if err != nil {
		fmt.Println("Error while opening output file: " + err.Error())
	}
	io.Copy(outFile, request.Body)
	saveFormData(outDir, request)
}

func getFileName(request *http.Request) string {
	fileNameHeader := request.Header.Get("X-File-Name")
	if fileNameHeader != "" {
		return fileNameHeader
	}
	return "data"
}
