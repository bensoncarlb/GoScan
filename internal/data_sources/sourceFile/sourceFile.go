// Data Souce module from File Pickups
package sourceFile

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	paramerror "github.com/bensoncb/GoScan/internal/errors"
	"github.com/bensoncb/GoScan/internal/structs/inputFile"
	"github.com/fsnotify/fsnotify"
	"github.com/rflandau/expiring"
)

type SourceConfig struct {
	Directory    string
	DataEndpoint string
	fsWatch      *fsnotify.Watcher
	chFiles      chan string
}

/*
* Monitor for new file events and filter out duplicates
* fsnotify doesn't expose an event for file close, so there are often duplicates
* due to the file system doing a write as multiple events.
 */
func filewatch(fsWatch *fsnotify.Watcher, chEvents chan string) {
	log.Printf("File watcher starting up")
	seenFiles := expiring.NewTable[string, bool]()
	timeout := time.Second * 5

	for {
		select {
		case event, ok := <-fsWatch.Events:
			if !ok {
				return
			}

			//Record seen files for a peroid of {timeout}
			//after not seeing the file for that period of time send it to the chan to be processed
			if event.Has(fsnotify.Write) {
				if _, found := seenFiles.Load(event.Name); !found {
					seenFiles.Store(event.Name, true, timeout, func(file string, _ bool) { chEvents <- file })
				} else {
					seenFiles.Refresh(event.Name, timeout)
				}
			}
		case event, ok := <-fsWatch.Errors:
			if !ok {
				return
			}
			panic(event.Error())
		}
	}
}

/*
* goroutine for handling new files on the specified (SourceConfig.Directory) location
 */
func fileevents(chFiles chan string, DataEndpoint string) {
	for {
		file, ok := <-chFiles

		if !ok {
			return
		}

		log.Printf("Received notification about new file: %s", file)

		//Give time for external file handlers to release
		time.Sleep(time.Millisecond * 500)

		f, err := os.ReadFile(file)

		if err != nil {
			panic(err)
		}

		i := inputFile.InputFile{Size: len(f), Name: filepath.Base(file), Src: "file", Data: f}

		b := new(bytes.Buffer)

		err = json.NewEncoder(b).Encode(i)
		if err != nil {
			panic(err)
		}

		//TODO Move to channel or direct?
		//Initially here as a HTTP call to allow for sourceFile.go to run from a separate system
		_, err = http.Post(DataEndpoint, "application/json", b)

		if err != nil {
			panic(err)
		}
	}
}

/*
* Start the FileWatch
 */
func (c *SourceConfig) Start() error {
	var err error
	c.fsWatch, err = fsnotify.NewBufferedWatcher(30)

	if err != nil {
		return err
	}
	log.Printf("Watching %s", c.Directory)
	if err := c.fsWatch.Add(c.Directory); err != nil {
		return err
	}

	c.chFiles = make(chan string, 50)

	go filewatch(c.fsWatch, c.chFiles)

	//Spawn worker
	for range 1 {
		go fileevents(c.chFiles, c.DataEndpoint)
	}

	return err
}

/*
* Cleanly stop the FileWatcher
 */
func (c *SourceConfig) Stop() error {
	c.fsWatch.Close()
	close(c.chFiles)
	return nil
}

/*
* Validate and prepare a SourceConfig to file pickup
 */
func New(Directory string, DataEndpoint string) (SourceConfig, error) {
	var err error

	if len(Directory) == 0 {
		err = paramerror.ErrBadParam{Parameter: "Directory", Reason: "Missing value"}
	} else if len(DataEndpoint) == 0 {
		err = paramerror.ErrBadParam{Parameter: "DataEndPoint", Reason: "Missing"}
	}

	if err == nil {
		_, err := os.Stat(Directory)

		if errors.Is(err, os.ErrNotExist) {
			err = os.Mkdir(Directory, os.ModePerm)
		}
	}

	conf := SourceConfig{Directory: Directory, DataEndpoint: DataEndpoint}

	return conf, err
}
