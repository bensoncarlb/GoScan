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
	var outputMethod string = "file" //Placeholder for switch below pending config support
	var outputDir string = "/home/carl/GoScan/rcvd"
	var inputDir string = "/home/carl/GoScan/test"

	ModOutput, err := outputFile.New(outputDir)

	if err != nil {
		panic(err)
	}

	switch outputMethod {
	case "file":
		//TODO
	default:
		panic(fmt.Errorf("Unrecognized output method %s", outputMethod))
	}

	//Setup listening server
	svr := server.Server{ModOutput: &ModOutput}

	if err := svr.Setup(); err != nil {
		panic(err)
	}

	if err := svr.Start(); err != nil {
		panic(err)
	}

	defer svr.Stop()

	//Setup DataInput listener
	DataInput, err := sourceFile.New(inputDir, "http://localhost:8090/data")

	if err != nil {
		panic(err)
	}

	if err := DataInput.Start(); err != nil {
		panic(err)
	}

	defer DataInput.Stop()

	//Wait for kill
	kill := make(chan os.Signal, 1)
	signal.Notify(kill, os.Interrupt)

	<-kill
}
