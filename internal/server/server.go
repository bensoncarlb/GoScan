// Listening server for new documents to process
package server

import (
	"bytes"
	"encoding/json"
	"image"
	"log"
	"maps"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/bensoncarlb/GoScan/internal/gsRecord"
	"github.com/bensoncarlb/GoScan/internal/ocr"
	"github.com/bensoncarlb/GoScan/internal/outputs/outputFile"
	"github.com/bensoncarlb/GoScan/structs"
)

type Server struct {
	ch                  chan gsRecord.RecordData
	isReady             bool
	httpServer          http.Server
	l                   net.Listener
	ModOutput           *outputFile.OutputModule
	DocumentTypes       map[string]structs.DocumentType
	DocumentLocation    string
	DocIdentifierRegion image.Rectangle
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
	sm.HandleFunc("/getitems", s.getItems)
	sm.HandleFunc("/retrieveitem", s.retrieveItem)
	sm.HandleFunc("/getdoctypes", s.getDocTypes)
	sm.HandleFunc("/adddoctype", s.addDocType)
	sm.HandleFunc("/deletedoctype", s.deleteDocType)

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
		go process(s.ch, s.ModOutput, s.DocIdentifierRegion, s.DocumentTypes)
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
func process(ch <-chan gsRecord.RecordData, outModule *outputFile.OutputModule, docTypeRegion image.Rectangle, documentTypes map[string]structs.DocumentType) {
	//Waiting for new item to process
	//TODO handle concurrency
	for outModule.IFile = range ch {
		log.Printf("Process routine received new item for processing: %s", outModule.IFile.Name)

		img := ocr.ConvertToGray(outModule.IFile.ImgData)

		docIdentifier, err := ocr.ReadRegion(img, docTypeRegion)

		if err != nil {
			log.Fatalf("Failed to get Document Type: %s", err)
		}

		docType, found := documentTypes[docIdentifier]

		if found {
			outModule.IFile.DocType = docType.Title
		} else {
			//TODO configurable
			outModule.IFile.DocType = "Default"
		}

		// Read and save off the document data via OCR
		if found && len(docType.Regions) > 0 {
			for _, docRegions := range docType.Regions {

				outModule.IFile.OCRData[docRegions.FieldName], err = ocr.ReadRegion(img, docRegions.Region)

				if err != nil {
					log.Fatalf("Failed to read image region %v for image %s", docRegions.RegionTitle, outModule.IFile.Name)
				}
			}
		} else {
			//If no regions are defined, read the entire image as a single field
			//TODO configurable
			outModule.IFile.OCRData["data"], err = ocr.ReadRegion(img, img.Bounds())

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

	d := gsRecord.RecordData{}
	err := json.NewDecoder(req.Body).Decode(&d)

	if err != nil {
		panic(err)
	}

	d.OCRData = map[string]string{}

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

func (s *Server) getItems(w http.ResponseWriter, req *http.Request) {
	items, err := s.ModOutput.List()

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
	}

	b := bytes.Buffer{}

	err = json.NewEncoder(&b).Encode(items)
	if err != nil {
		panic(err)
		//TODO better handling
	}

	w.Write(b.Bytes())
}

func (s *Server) retrieveItem(w http.ResponseWriter, req *http.Request) {
	itemReq := structs.ReqRetrieveItem{}
	if err := json.NewDecoder(req.Body).Decode(&itemReq); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	} else if strings.TrimSpace(itemReq.ItemName) == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	record, err := s.ModOutput.GetItem(itemReq.ItemName)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	rsp := structs.RspRetrieveItem{Fields: record.OCRData, ImgData: record.ImgData}

	b := bytes.Buffer{}
	json.NewEncoder(&b).Encode(rsp)

	w.Write(b.Bytes())
}

func (s *Server) getDocTypes(w http.ResponseWriter, req *http.Request) {
	rsp := structs.RspGetDocumentTypes{DocumentTypes: slices.Collect(maps.Values(s.DocumentTypes))}

	b := bytes.Buffer{}

	if err := json.NewEncoder(&b).Encode(rsp); err != nil {
		panic(err)
		//TODO better handling
	}

	w.Write(b.Bytes())
}

func (s *Server) deleteDocType(w http.ResponseWriter, req *http.Request) {
	d := bytes.NewBuffer
	err := json.NewDecoder(req.Body).Decode(&d)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	//TODO sort out parsing the incoming request
	if docType, ok := s.DocumentTypes[""]; !ok {
		w.WriteHeader(http.StatusBadRequest)
	} else {
		delete(s.DocumentTypes, docType.Title)
		os.Remove(docType.Identifier)

		w.WriteHeader(http.StatusOK)
	}
}

func (s *Server) addDocType(w http.ResponseWriter, req *http.Request) {
	d := structs.DocumentType{}
	err := json.NewDecoder(req.Body).Decode(&d)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	} else if !d.IsValid() {
		w.WriteHeader(http.StatusBadRequest)
		return
	} else if _, ok := s.DocumentTypes[d.Identifier]; ok {
		//Already exists
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(s.DocumentLocation) == "" {
		w.WriteHeader(http.StatusInternalServerError)
		//TODO better logging
		return
	}

	fil, err := os.Create(filepath.Join(s.DocumentLocation, d.Identifier))

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		//TODO better logging
		return
	}

	defer fil.Close()

	b := bytes.Buffer{}
	json.NewEncoder(&b).Encode(d)

	fil.Write(b.Bytes())
	s.DocumentTypes[d.Identifier] = d

	w.WriteHeader(http.StatusAccepted)
}
