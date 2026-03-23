// Listening server for new documents to process
package server

import (
	_ "bytes"
	"encoding/json"
	_ "image"
	"log"
	"net"
	"net/http"

	"github.com/bensoncb/GoScan/internal/gsRecord"
	"github.com/bensoncb/GoScan/internal/ocr"
	"github.com/bensoncb/GoScan/internal/outputs/outputFile"
)

type Server struct {
	ch         chan gsRecord.RecordData
	isReady    bool
	httpServer http.Server
	l          net.Listener
	ModOutput  *outputFile.OutputModule
}

// Setup the listening server
func (s *Server) New() error {
	//Setup a channel for processing incoming input files
	//TODO configurable
	s.ch = make(chan gsRecord.RecordData, 50)

	s.httpServer = http.Server{}
	sm := http.NewServeMux()

	sm.HandleFunc("/data", s.receiveData)
	sm.HandleFunc("/ping", s.ping)

	s.httpServer.Handler = sm

	return nil
}

// Startup the listening server
func (s *Server) Start() error {
	log.Printf("Data Server starting up")

	if s.isReady {
		return nil
	}

	var err error

	s.l, err = net.Listen("tcp", ":8090") //TODO configurable port

	if err != nil {
		panic(err)
	}

	//TODO configurable
	for range 1 {
		//Kick off routine(s) to listen for new items to process
		go process(s.ch, s.ModOutput)
	}

	go s.httpServer.Serve(s.l)

	s.isReady = true

	return nil
}

//TODO stop vs terminate

// Cleanly stop the listening server
func (s *Server) Stop() error {
	s.isReady = false

	if s.l != nil {
		s.l.Close()
	}

	s.httpServer.Close()

	if s.ch != nil {
		close(s.ch)
		s.ch = nil
	}

	return nil
}

// Func for goroutines to process incoming submissions to /data
func process(ch <-chan gsRecord.RecordData, outModule *outputFile.OutputModule) {
	ok := true

	for {
		//Block waiting for new item to process
		outModule.IFile, ok = <-ch
		if !ok {
			return
		}

		log.Printf("Process routine received new item for processing: %s", outModule.IFile.Name)

		//TODO Check DecodeConfig
		var err error
		/*_, _, err := image.Decode(bytes.NewReader(outModule.IFile.ImgData))

		if err != nil {
			log.Fatalf("Failed reading image: %s", err)
		}
		*/
		/*
			buf := new(bytes.Buffer)
			_ = png.Encode(buf, imgBase)
			b := buf.Bytes()
		*/

		// Read and save off the document data via OCR
		if len(outModule.IFile.Regions) > 0 {
			/*for field, reg := range outModule.IFile.Regions {
				buf := new(bytes.Buffer)

				err := png.Encode(buf, img.SubImage(reg))

				if err != nil {
					log.Fatalf("Failed to read image region %v for image %s", reg, outModule.IFile.Name)
				}

				regImgData := buf.Bytes()

				outModule.IFile.OCRData[field], err = ocr.ReadRegion(&regImgData)

				if err != nil {
					log.Fatalf("Failed to read image region %v for image %s", reg, outModule.IFile.Name)
				}
			}*/
		} else {
			//If no regions are defined, read the entire image as a single field
			//TODO configurable
			outModule.IFile.OCRData["data"], err = ocr.ReadRegion(outModule.IFile.ImgData)

			if err != nil {
				log.Fatalf("Failed to read data for %s", outModule.IFile.Name)
			}
		}

		if err != nil {
			panic(err)
		}

		// Save off the incoming data via the Output Module
		if err := outModule.Save(); err != nil {
			panic(err)
		}

	}
}

// func handler for /data endpoint
func (dr *Server) receiveData(w http.ResponseWriter, req *http.Request) {
	log.Printf("Received new request from: %s", req.RemoteAddr)

	d := gsRecord.RecordData{OCRData: make(map[string]string)}
	err := json.NewDecoder(req.Body).Decode(&d)

	if err != nil {
		panic(err)
	}

	dr.ch <- d
	w.WriteHeader(http.StatusAccepted)
}

// func handler for /ping endpoint
func (dr *Server) ping(w http.ResponseWriter, req *http.Request) {
	if dr.isReady {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
