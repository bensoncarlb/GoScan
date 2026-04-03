// Listening server for new documents to process
package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	_ "image/png"
	"log"
	"maps"
	"net"
	"net/http"
	"os"
	"slices"
	"strings"

	"github.com/bensoncarlb/GoScan/internal/gsRecord"
	"github.com/bensoncarlb/GoScan/internal/gserrors"
	"github.com/bensoncarlb/GoScan/internal/ocr"
	"github.com/bensoncarlb/GoScan/internal/outputs"
	"github.com/bensoncarlb/GoScan/structs"
)

type Server struct {
	chPendingDocs       chan gsRecord.RecordData
	isReady             bool
	httpServer          http.Server
	l                   net.Listener
	ModOutput           outputs.Module
	DocumentTypes       map[string]structs.DocumentType
	DocumentLocation    string
	DocumentTypeRoot    os.Root
	DocIdentifierRegion image.Rectangle
}

// TODO make constructor rather than method
// Setup the listening server
func New(modOutput outputs.Module, docIdentifierRegion image.Rectangle, docTypeDir string) (*Server, error) {
	//Setup a channel for processing incoming input files
	//TODO configurable
	svr := Server{
		ModOutput:           modOutput,
		DocIdentifierRegion: image.Rect(1200, 1800, 1700, 2200),
		DocumentLocation:    docTypeDir}

	docRoot, err := os.OpenRoot(docTypeDir)

	if err != nil {
		return &Server{}, err
	}

	svr.DocumentTypeRoot = *docRoot

	svr.DocumentTypes, err = LoadDocumentTypes(docTypeDir)

	if err != nil {
		return &Server{}, err
	}

	svr.chPendingDocs = make(chan gsRecord.RecordData, 50)

	svr.httpServer = http.Server{}
	sm := http.NewServeMux()

	sm.HandleFunc("/data", svr.receiveData)
	sm.HandleFunc("/ping", svr.ping)
	sm.HandleFunc("/getitems", svr.getItems)
	sm.HandleFunc("/retrieveitem", svr.retrieveItem)
	sm.HandleFunc("/getdoctypes", svr.getDocTypes)
	sm.HandleFunc("/adddoctype", svr.addDocType)
	sm.HandleFunc("/deletedoctype", svr.deleteDocType)

	svr.httpServer.Handler = sm

	return &svr, nil
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
		go process(s.chPendingDocs, s.ModOutput, s.DocIdentifierRegion, s.DocumentTypes)
	}

	go s.httpServer.Serve(s.l)

	fmt.Println("Listening on localhost:8090")

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

	if s.chPendingDocs != nil {
		close(s.chPendingDocs)
		s.chPendingDocs = nil
	}

	return nil
}

func checkRectangle(rImage image.Rectangle, rRegion image.Rectangle) bool {
	if rImage.Min.X < rRegion.Min.X || rImage.Min.Y < rRegion.Min.Y {
		return false
	} else if rImage.Max.X < rRegion.Max.X || rImage.Max.Y < rRegion.Max.Y {
		return false
	}

	return true
}

func LoadDocumentTypes(directory string) (map[string]structs.DocumentType, error) {
	if strings.TrimSpace(directory) == "" {
		return nil, gserrors.ErrBadParam{Parameter: "Directory", Reason: "Missing"}
	} else if _, err := os.Stat(directory); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			//If no matching directory exists, create it
			err = os.MkdirAll(directory, os.ModePerm)

			if err != nil {
				return nil, err
			}

			//Since it didn't exist no document types to return
			return map[string]structs.DocumentType{}, nil
		} else {
			return nil, err
		}
	}

	types, err := os.ReadDir(directory)

	if err != nil {
		return nil, err
	}

	docTypes := make(map[string]structs.DocumentType, len(types))

	for _, dirEntry := range types {
		if dirEntry.IsDir() {
			continue
		}

		f, err := os.OpenInRoot(directory, dirEntry.Name())

		if err != nil {
			return nil, err
		}

		d := structs.DocumentType{}
		err = json.NewDecoder(f).Decode(&d)

		if err != nil {
			return nil, err
		}

		docTypes[d.Identifier] = d
	}

	return docTypes, nil
}

// Func for goroutines to process incoming submissions to /data
func process(ch <-chan gsRecord.RecordData, outModule outputs.Module, docTypeRegion image.Rectangle, documentTypes map[string]structs.DocumentType) {
	//Waiting for new item to process
	for newRecord := range ch {
		log.Printf("Process routine received new item for processing: %s", newRecord.Name)

		img := ocr.ConvertToGray(newRecord.ImgData)
		var docType structs.DocumentType
		var err error

		if checkRectangle(img.Bounds(), docTypeRegion) {
			docIdentifier, err := ocr.ReadRegion(img, docTypeRegion)

			if err != nil {
				log.Fatalf("Failed to get Document Type: %s", err)
			}

			docType, found := documentTypes[strings.ToLower(docIdentifier[:8])]

			if found {
				newRecord.DocType = docType.Title
			} else {
				//TODO configurable
				newRecord.DocType = "Default"
			}
		} else {
			newRecord.DocType = "Default"
		}

		// Read and save off the document data via OCR
		if len(docType.Regions) > 0 {
			for _, docRegions := range docType.Regions {

				newRecord.OCRData[docRegions.FieldName], err = ocr.ReadRegion(img, docRegions.Region)

				if err != nil {
					log.Fatalf("Failed to read image region %v for image %s", docRegions.RegionTitle, newRecord.Name)
				}

				newRecord.OCRData[docRegions.FieldName] = strings.TrimRight(newRecord.OCRData[docRegions.FieldName], "\n")
			}
		} else {
			//If no regions are defined, read the entire image as a single field
			//TODO configurable
			newRecord.OCRData["data"], err = ocr.ReadRegion(img, img.Bounds())

			if err != nil {
				log.Fatalf("Failed to read data: %s", err)
			}
		}

		if err != nil {
			panic(err)
		}

		// Save off the incoming data via the Output Module
		if err := outModule.Save(&newRecord); err != nil {
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

	dr.chPendingDocs <- d
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
	items, err := s.ModOutput.ListItems()

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

	record, err := s.ModOutput.Retrieve(itemReq.ItemName)
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
	d := structs.ReqDeleteDocumentType{}
	err := json.NewDecoder(req.Body).Decode(&d)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if _, ok := s.DocumentTypes[d.DocumentType]; !ok {
		w.WriteHeader(http.StatusBadRequest)
	} else {
		delete(s.DocumentTypes, d.DocumentType)
		s.DocumentTypeRoot.Remove(d.DocumentType)

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

	fil, err := s.DocumentTypeRoot.Open(d.Identifier)

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
