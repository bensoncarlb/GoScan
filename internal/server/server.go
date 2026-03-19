package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/bensoncb/GoScan/internal/ocr"
	"github.com/bensoncb/GoScan/internal/outputs/outputFile"
	"github.com/bensoncb/GoScan/internal/structs/inputFile"
)

type Server struct {
	ch         chan inputFile.InputFile
	isReady    bool
	httpServer http.Server
	l          net.Listener
	ModOutput  *outputFile.OutputModule
}

func (s *Server) Setup() error {
	log.Println("Data Server Starting up")

	//Setup a channel for processing incoming input files
	s.ch = make(chan inputFile.InputFile)

	s.httpServer = http.Server{}
	sm := http.NewServeMux()

	sm.HandleFunc("/data", s.data)
	sm.HandleFunc("/ping", s.ping)

	s.httpServer.Handler = sm

	return nil
}

func (s *Server) Start() error {
	var err error

	s.l, err = net.Listen("tcp", ":8090") //TODO Port assignment

	if err != nil {
		panic(err)
	}

	for range 1 {
		//Kick off routine(s) to listen for new items to process
		go process(s.ch, s.ModOutput)
	}

	go s.httpServer.Serve(s.l)

	s.isReady = true

	return nil
}

func (s *Server) Stop() error {
	s.l.Close()
	s.httpServer.Close()
	close(s.ch)

	return nil
}
func process(ch <-chan inputFile.InputFile, outModule *outputFile.OutputModule) {
	ok := true

	for {
		//Block waiting for new item to process
		outModule.IFile, ok = <-ch
		if !ok {
			return
		}

		log.Println("Process routine received new item for processing: ", outModule.IFile.Name)

		if err := outModule.Save(); err != nil {
			panic(err)
		}

		//"Read" the incoming item for indexing data
		ocrData, err := ocr.ReadImage(&outModule.IFile.Data)

		if err != nil {
			panic(err)
		}

		//Print off results
		println(ocrData)

		fmt.Println(outModule.IFile)
	}
}

func (dr *Server) data(w http.ResponseWriter, req *http.Request) {
	log.Println("Received new request", req.RemoteAddr)

	d := inputFile.InputFile{}
	err := json.NewDecoder(req.Body).Decode(&d)

	if err != nil {
		panic(err)
	}

	dr.ch <- d
	w.WriteHeader(http.StatusAccepted)
}

func (dr *Server) ping(w http.ResponseWriter, req *http.Request) {
	if dr.isReady {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
