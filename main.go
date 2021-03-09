package main

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/julienschmidt/httprouter"
)

func main() {
	router := httprouter.New()
	router.PUT("/upload", receiveImage)
	err := http.ListenAndServe("0.0.0.0:3000", router)
	if err != nil {
		// log.Fatal(err)
		fmt.Println("error")
	}
}

func receiveImage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	err := uploadImage(r)
	if err != nil {
		http.Error(w, "Invalid Data", http.StatusBadRequest)
		return
	}
}

func uploadImage(r *http.Request) error {
	r.ParseMultipartForm(32 << 20)

	file, handler, err := r.FormFile("file")
	if err != nil {
		return err
	}
	defer file.Close()

	f, err := os.OpenFile("/Users/scott/dev/guppy/"+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer f.Close()

	io.Copy(f, file)
	return nil
}
