// Primary package for coordinating processing server (server.go), data pickup (data_sources/), and data export (outputs/)
package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/bensoncb/GoScan/internal/data_sources/sourceFile"
	"github.com/bensoncb/GoScan/internal/outputs/outputFile"
	"github.com/bensoncb/GoScan/internal/server"
)

func main() {
	//Setup handler for outputing final data
	//TODO configurable
	var outputMethod string = "file" //Placeholder for switch below pending config support
	var outputDir string = "/home/carl/GoScan/rcvd"
	var inputDir string = "/home/carl/GoScan/test"

	/***
	* Setup the output module to be passed to the server.go process
	***/
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

	/***
	* Setup listening server
	***/
	svr := server.Server{ModOutput: &modOutput}

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

	/***
	* Setup the data source listener module
	***/
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

	/***
	* Wait for kill signal
	***/
	kill := make(chan os.Signal, 1)
	signal.Notify(kill, os.Interrupt)

	<-kill
}
