package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	gs_s "github.com/bensoncb/GoScan/structs"
	"github.com/fsnotify/fsnotify"
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

func filewatcher(fsWatch *fsnotify.Watcher) {
	log.Println("File watcher starting up")
	for {
		select {
		case event, ok := <-fsWatch.Events:
			if !ok {
				return
			}

			if event.Has(fsnotify.Write) {
				log.Println("Received notification about new file:", event.Name)
				time.Sleep(time.Millisecond * 500)

				f, err := os.ReadFile(event.Name)

				if err != nil {
					panic(err)
				}

				i := gs_s.InputFile{Size: len(f), Name: event.Name[5:], Src: "file", Data: f}

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
}

func main() {
	//Set up watcher
	fsWatch, err := fsnotify.NewBufferedWatcher(30)
	if err != nil {
		panic(err)
	}

	defer fsWatch.Close()

	DirCheck("test")

	if err := fsWatch.Add("test"); err != nil {
		panic(err)
	}

	//Spawn worker
	for range 1 {
		go filewatcher(fsWatch)
	}

	//Wait for kill
	kill := make(chan os.Signal, 1)
	signal.Notify(kill, os.Interrupt)

	<-kill

	//Close channels
	fsWatch.Close()
}
