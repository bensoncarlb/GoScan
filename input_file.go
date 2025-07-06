package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/fsnotify/fsnotify"
)

type InputFile struct {
	size      int
	name, src string
}

func main() {
	//Set up watcher
	fsWatch, err := fsnotify.NewBufferedWatcher(30)
	if err != nil {
		panic(err)
	}

	defer fsWatch.Close()

	//Add directory
	if err := fsWatch.Add("."); err != nil {
		panic(err)
	}

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
						println("1")
						//time.Sleep(time.Millisecond * 500)

						f, err := os.ReadFile(event.Name)

						if err != nil {
							panic(err)
						}

						i := InputFile{size: len(f), name: event.Name, src: "file"}

						fmt.Println(i)
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
