package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/fsnotify/fsnotify"
)

type InputFile struct {
	Size int    `json:"size"`
	Name string `json:"name"`
	Src  string `json:"src"`
	Data []byte `json:"data"`
}

func data(w http.ResponseWriter, req *http.Request) {
	d := &InputFile{}
	err := json.NewDecoder(req.Body).Decode(d)

	if err != nil {
		panic(err)
	}

	fmt.Printf("%+v\n", d)
	w.WriteHeader(http.StatusCreated)

	f, err := os.Create("rcvd" + d.Name)

	if err != nil {
		panic(err)
	}

	f.Write(d.Data)
	f.Close()
}

func main() {
	//Set up watcher
	fsWatch, err := fsnotify.NewBufferedWatcher(30)
	if err != nil {
		panic(err)
	}

	defer fsWatch.Close()

	//Add directory
	if err := fsWatch.Add("test"); err != nil {
		panic(err)
	}

	go func() {
		http.HandleFunc("/data", data)

		http.ListenAndServe(":8090", nil)
	}()

	//Spawn worker
	for range 1 {
		go func() {
			for {
				select {
				case event, ok := <-fsWatch.Events:
					if !ok {
						return
					}

					if event.Has(fsnotify.Write) {
						time.Sleep(time.Millisecond * 500)

						f, err := os.ReadFile(event.Name)

						if err != nil {
							panic(err)
						}

						i := InputFile{Size: len(f), Name: event.Name[5:], Src: "file", Data: f}

						b := new(bytes.Buffer)

						err = json.NewEncoder(b).Encode(i)
						if err != nil {
							panic(err)
						}

						resp, err := http.Post("http://localhost:8090/data", "application/json", b)

						if err != nil {
							panic(err)
						}

						defer resp.Body.Close()

						fmt.Println(resp.Status)
					}
				case event, ok := <-fsWatch.Errors:
					if !ok {
						return
					}
					panic(event.Error())
				}
			}
		}()
	}

	//Wait for kill
	kill := make(chan os.Signal, 1)
	signal.Notify(kill, os.Interrupt)

	<-kill

	//Close channels
	fsWatch.Close()
}
