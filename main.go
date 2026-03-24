// Primary package for coordinating processing server (server.go), data pickup (data_sources/), and data export (outputs/)
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"os"
	"os/signal"
	"strings"

	"github.com/bensoncb/GoScan/internal/data_sources/sourceFile"
	"github.com/bensoncb/GoScan/internal/documentType"
	"github.com/bensoncb/GoScan/internal/gserrors"
	"github.com/bensoncb/GoScan/internal/outputs/outputFile"
	"github.com/bensoncb/GoScan/internal/server"
)

func main() {
	//Setup handler for outputing final data
	//TODO configurable
	var outputMethod string = "file" //Placeholder for switch below pending config support
	var outputDir string = "/home/carl/GoScan/rcvd"
	var inputDir string = "/home/carl/GoScan/test"
	var docTypeDir string = "/home/carl/GoScan/DocumentTypes"

	//
	// Setup the output module to be passed to the server.go process
	//
	modOutput, err := outputFile.New(outputDir)

	if err != nil {
		panic(err)
	}

	switch outputMethod {
	case "file":
		//TODO Something with output method
	default:
		panic(fmt.Errorf("unrecognized output method %s", outputMethod))
	}

	//
	// Load configured document types
	//
	docTypes, err := LoadDocumentTypes(docTypeDir)

	if err != nil {
		panic(err)
	}
	//
	// Setup listening server
	//
	//TODO setup identifier region
	svr := server.Server{ModOutput: &modOutput, DocumentTypes: docTypes, DocIdentifierRegion: image.Rect(0, 0, 1, 1)}

	err = svr.New()
	if err != nil {
		panic(err)
	}

	//Start up the server
	err = svr.Start()
	if err != nil {
		panic(err)
	}

	defer svr.Stop()

	//
	// Setup the data source listener module
	//
	dataInput, err := sourceFile.New(inputDir, "http://localhost:8090/data") //TODO configurable

	if err != nil {
		panic(err)
	}

	//Start the data source listener
	err = dataInput.Start()

	if err != nil {
		panic(err)
	}

	defer dataInput.Stop()

	//
	// Wait for kill signal
	//
	kill := make(chan os.Signal, 1)
	signal.Notify(kill, os.Interrupt)

	<-kill
}

func LoadDocumentTypes(directory string) (map[string]documentType.DocumentType, error) {
	if strings.TrimSpace(directory) == "" {
		return nil, gserrors.ErrBadParam{Parameter: "Directory", Reason: "Missing"}
	} else if _, err := os.Stat(directory); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			err = os.MkdirAll(directory, os.ModePerm)

			if err != nil {
				return nil, err
			}

			//If no matching directory exists, create it
			//Since it didn't exist no document types to return
			return map[string]documentType.DocumentType{}, nil
		} else {
			return nil, err
		}
	}

	types, err := os.ReadDir(directory)

	if err != nil {
		return nil, err
	}

	docTypes := make(map[string]documentType.DocumentType, len(types))

	for _, dirEntry := range types {
		if dirEntry.IsDir() {
			continue
		}

		f, err := os.Open(dirEntry.Name())

		if err != nil {
			return nil, err
		}

		d := documentType.DocumentType{}
		err = json.NewDecoder(f).Decode(&d)

		if err != nil {
			return nil, err
		}

		docTypes[d.Identifier] = d
	}

	return docTypes, nil
}
