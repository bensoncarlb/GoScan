package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/bensoncb/GoScan/internal/ocr"
	"github.com/bensoncb/GoScan/internal/structs"
)

func data(w http.ResponseWriter, req *http.Request) {
	log.Println("Received new request", req.RemoteAddr)

	d := &structs.InputFile{}
	err := json.NewDecoder(req.Body).Decode(d)

	if err != nil {
		panic(err)
	}

	fmt.Printf("%+v\n", d)
	w.WriteHeader(http.StatusCreated)

	f, err := os.Create("rcvd/" + d.Name)

	if err != nil {
		panic(err)
	}

	defer f.Close()

	f.Write(d.Data)

	s, err := ocr.ReadImage(&d.Data)

	if err != nil {
		panic(err)
	}

	println(s)
}

func main() {
	_, err := os.Stat("rcvd")

	if errors.Is(err, os.ErrNotExist) {
		err = os.Mkdir("rcvd", os.ModePerm)

		if err != nil {
			panic(err)
		}
	}

	http.HandleFunc("/data", data)

	log.Println("Data Server Starting up")

	go http.ListenAndServe(":8090", nil)

	//Wait for kill
	kill := make(chan os.Signal, 1)
	signal.Notify(kill, os.Interrupt)

	<-kill
}
