package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"

	gs_s "github.com/bensoncb/GoScan/structs"
)

func DirCheck(p string) {
	_, err := os.Stat(p)

	if errors.Is(err, os.ErrNotExist) {
		err = os.Mkdir(p, os.ModePerm)

		if err != nil {
			panic(err)
		}
	}
}

func data(w http.ResponseWriter, req *http.Request) {
	log.Println("Received new request", req.RemoteAddr)

	d := &gs_s.InputFile{}
	err := json.NewDecoder(req.Body).Decode(d)

	if err != nil {
		panic(err)
	}

	fmt.Printf("%+v\n", d)
	w.WriteHeader(http.StatusCreated)

	f, err := os.Create("rcvd/" + d.Name)

	if err != nil {
		panic(err)
	}

	f.Write(d.Data)
	f.Close()
}

func httpwatcher() {
	log.Println("Data Server Starting up")
	DirCheck("rcvd")
	http.ListenAndServe(":8090", nil)
}

func main() {
	http.HandleFunc("/data", data)

	go httpwatcher()

	//Wait for kill
	kill := make(chan os.Signal, 1)
	signal.Notify(kill, os.Interrupt)

	<-kill
}
