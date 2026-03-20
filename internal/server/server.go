// Listening server for new documents to process
package server

import (
	"encoding/json"
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

/*
* Setup the listening server
 */
func (s *Server) Setup() error {
	log.Printf("Data Server starting up")

	//Setup a channel for processing incoming input files
	s.ch = make(chan inputFile.InputFile)

	s.httpServer = http.Server{}
	sm := http.NewServeMux()

	sm.HandleFunc("/data", s.data)
	sm.HandleFunc("/ping", s.ping)

	s.httpServer.Handler = sm

	return nil
}

/*
* Startup the listening server
 */
func (s *Server) Start() error {
	if s.isReady {
		return nil
	}

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

/*
* Cleanly stop the listening server
 */
func (s *Server) Stop() error {
	s.isReady = false

	s.l.Close()
	s.httpServer.Close()
	close(s.ch)

	return nil
}

/*
* Func for goroutines to process incoming submissions to /data
 */
func process(ch <-chan inputFile.InputFile, outModule *outputFile.OutputModule) {
	ok := true

	for {
		//Block waiting for new item to process
		outModule.IFile, ok = <-ch
		if !ok {
			return
		}

		log.Printf("Process routine received new item for processing: %s", outModule.IFile.Name)
		// Read and save off the document data via OCR
		_, err := ocr.ReadImage(&outModule.IFile.Data)
		//TODO Save off data in the inputfile

		if err != nil {
			panic(err)
		}

		// Save off the incoming data via the Output Module
		if err := outModule.Save(); err != nil {
			panic(err)
		}

	}
}

/*
* func handler for /data endpoint
 */
func (dr *Server) data(w http.ResponseWriter, req *http.Request) {
	log.Printf("Received new request from: %s", req.RemoteAddr)

	d := inputFile.InputFile{}
	err := json.NewDecoder(req.Body).Decode(&d)

	if err != nil {
		panic(err)
	}

	dr.ch <- d
	w.WriteHeader(http.StatusAccepted)
}

/*
* func handler for /ping endpoint
 */
func (dr *Server) ping(w http.ResponseWriter, req *http.Request) {
	if dr.isReady {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
