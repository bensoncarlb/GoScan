// Primary package for coordinating processing server (server.go), data pickup (data_sources/), and data export (outputs/)
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"os"
	"os/signal"
	"path/filepath"
	"strings"

	"github.com/bensoncarlb/GoScan/internal/gserrors"
	"github.com/bensoncarlb/GoScan/internal/inputs"
	"github.com/bensoncarlb/GoScan/internal/outputs"
	"github.com/bensoncarlb/GoScan/internal/server"
	"github.com/bensoncarlb/GoScan/structs"
)

func main() {
	//Setup handler for outputing final data
	//TODO configurable
	var outputMethod = "file" //Placeholder for switch below pending config support
	path, err := os.Getwd()

	if err != nil {
		panic(err)
	}
	var outputDir = filepath.Join(path, "output")
	var inputDir = filepath.Join(path, "pickup")
	inputMethod := "file"
	var docTypeDir = filepath.Join(path, "DocumentTypes")

	//
	// Setup the output module to be passed to the server.go process
	//
	var modOutput outputs.Module

	switch strings.ToLower(outputMethod) {
	case "file":
		modOutput = outputs.OutputFile{Directory: outputDir}
		err = modOutput.Init()
	default:
		panic(fmt.Errorf("unrecognized output method %s", outputMethod))
	}

	if err != nil {
		panic(err)
	}

	//
	// Load configured document types
	//
	//TODO move to server.go
	docTypes, err := LoadDocumentTypes(docTypeDir)

	if err != nil {
		panic(err)
	}
	//
	// Setup listening server
	//
	//TODO setup identifier region
	svr, err := server.New(
		modOutput,
		docTypes,
		image.Rect(1200, 1800, 1700, 2200),
		docTypeDir)

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
	var dataInput inputs.Module

	switch strings.ToLower(inputMethod) {
	case "file":
		dataInput = inputs.FileWatch{Directory: inputDir, DataEndpoint: "http://localhost:8090/data"}
		err = dataInput.Init()
	default:
		panic("unknown input method: " + inputMethod)
	}

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
