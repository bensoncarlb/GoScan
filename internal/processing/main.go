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

type data_rcvd struct {
	ch chan structs.InputFile
}

func process(ch <-chan structs.InputFile) {
	for {
		//Block waiting for new item to process
		d := <-ch

		log.Println("Process routine received new item for processing: ", d.Name)

		//Save off received data
		f, err := os.Create("rcvd/" + d.Name)

		if err != nil {
			panic(err)
		}

		f.Write(d.Data)
		f.Close()

		//"Read" the incoming item for indexing data
		s, err := ocr.ReadImage(&d.Data)

		if err != nil {
			panic(err)
		}

		//Print off results
		println(s)

		fmt.Println(d)
	}
}

func (dr *data_rcvd) data(w http.ResponseWriter, req *http.Request) {
	log.Println("Received new request", req.RemoteAddr)

	d := &structs.InputFile{}
	err := json.NewDecoder(req.Body).Decode(d)

	if err != nil {
		panic(err)
	}

	w.WriteHeader(http.StatusCreated)

	dr.ch <- *d
}

func main() {
	log.Println("Data Server Starting up")

	//Check the directory to (for now) store received data in exists
	_, err := os.Stat("rcvd")

	if errors.Is(err, os.ErrNotExist) {
		err = os.Mkdir("rcvd", os.ModePerm)

		if err != nil {
			panic(err)
		}
	}

	//Setup a channel for processing incoming input files
	ch_process := &data_rcvd{ch: make(chan structs.InputFile)}

	for range 1 {
		//Kick off routine(s) to listen for new items to process
		go process(ch_process.ch)
	}

	//Setup HTTP listener and handler
	http.HandleFunc("/data", ch_process.data)

	go http.ListenAndServe(":8090", nil)

	//Wait for kill
	kill := make(chan os.Signal, 1)
	signal.Notify(kill, os.Interrupt)

	<-kill
}
