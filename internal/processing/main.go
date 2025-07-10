package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"net"

	"github.com/bensoncb/GoScan/internal/ocr"
	"github.com/bensoncb/GoScan/internal/structs/inputFile"
)

type server struct {
	ch      chan inputFile.InputFile
	isReady bool
	httpServer 	http.Server
	l net.Listener
}

func process(ch <-chan inputFile.InputFile) {
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

func (dr *server) data(w http.ResponseWriter, req *http.Request) {
	log.Println("Received new request", req.RemoteAddr)

	d := inputFile.InputFile{}
	err := json.NewDecoder(req.Body).Decode(&d)

	if err != nil {
		panic(err)
	}

	dr.ch <- d
	w.WriteHeader(http.StatusAccepted)
}

func (dr *server) ping(w http.ResponseWriter, req *http.Request) {
	if dr.isReady {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func main() {
	log.Println("Data Server Starting up")

	//Check the directory to (for now) store received data in exists
	if fi, err := os.Stat("rcvd"); errors.Is(err, os.ErrNotExist) {
		err = os.Mkdir("rcvd", os.ModePerm)

		if err != nil {
			panic(err)
		}
	} else if !fi.IsDir() {
		panic("Dir exists as a file")
	}

	//Setup a channel for processing incoming input files
	svr := &server{ch: make(chan inputFile.InputFile)}

	for range 1 {
		//Kick off routine(s) to listen for new items to process
		go process(svr.ch)
	}
	
	var err error
	svr.l, err = net.Listen("tcp", ":8090")

	if err != nil{
		panic(err)
	}

	defer svr.l.Close()

	svr.httpServer = http.Server{}
	sm := http.NewServeMux()

	sm.HandleFunc("/data", svr.data)
	sm.HandleFunc("/ping", svr.ping)

	svr.httpServer.Handler = sm

	go svr.httpServer.Serve(svr.l)

	svr.isReady = true

	//Wait for kill
	kill := make(chan os.Signal, 1)
	signal.Notify(kill, os.Interrupt)

	<-kill
}
