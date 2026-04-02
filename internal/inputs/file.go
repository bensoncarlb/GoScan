package inputs

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bensoncarlb/GoScan/internal/gsRecord"
	"github.com/bensoncarlb/GoScan/internal/gserrors"
	"github.com/fsnotify/fsnotify"
	"github.com/rflandau/expiring"
)

// TODO rename to a ~FileConfig
type FileWatch struct {
	isRunning    bool
	Directory    string
	DataEndpoint string
	fsWatch      *fsnotify.Watcher
	chFiles      chan string
}

// Monitor for new file events and filter out duplicates
// fsnotify doesn't expose an event for file close, so there are often duplicates
// due to the file system doing a write as multiple events.
func fileWatch(fsWatch *fsnotify.Watcher, chEvents chan string) {
	log.Printf("File watcher starting up")
	seenFiles := expiring.NewTable[string, bool]()

	//TODO configurable
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
					log.Println("New file: " + event.Name)
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

// fileEvents handles processing new files picked up from the specified (file.Directory) location.
func fileEvents(chFiles <-chan string, DataEndpoint string) {
	for file := range chFiles {
		log.Printf("Received notification about new file: %s", file)

		f, err := os.ReadFile(file)

		if err != nil {
			panic(err)
		}

		//TODO enumerate src
		record, err := gsRecord.New(len(f), filepath.Base(file), "file", f)

		b := bytes.Buffer{}

		err = json.NewEncoder(&b).Encode(record)
		if err != nil {
			panic(err)
		}

		//TODO Move to channel or direct?
		//Initially here as a HTTP call to allow for sourceFile.go to run from a separate system
		_, err = http.Post(DataEndpoint, "application/json", &b)

		if err != nil {
			panic(err)
		}

		//Clean up file
		err = os.Remove(file)

		if err != nil {
			panic(err)
		}
	}
}

// Start the FileWatch
func (c FileWatch) Start() error {
	if c.isRunning {
		return nil
	}

	var err error
	//TODO configurable
	c.fsWatch, err = fsnotify.NewBufferedWatcher(30)

	if err != nil {
		return err
	}

	if err := c.fsWatch.Add(c.Directory); err != nil {
		return err
	}

	log.Printf("Watching %s", c.Directory)

	//TODO configurable
	c.chFiles = make(chan string, 50)

	go fileWatch(c.fsWatch, c.chFiles)

	//Spawn worker
	//TODO configurable
	for range 1 {
		go fileEvents(c.chFiles, c.DataEndpoint)
	}

	existingFiles, err := os.ReadDir(c.Directory)

	if err != nil {
		c.Stop()
		return err
	}

	for _, f := range existingFiles {
		//TODO check Type instead
		if !f.IsDir() {
			c.chFiles <- filepath.Join(c.Directory, f.Name())
		}
	}

	c.isRunning = true

	return nil
}

// Cleanly stop the FileWatcher
func (c FileWatch) Stop() error {
	if !c.isRunning {
		return nil
	}

	if c.fsWatch != nil {
		c.fsWatch.Close()
	}

	if c.chFiles != nil {
		close(c.chFiles)
		c.chFiles = nil
	}

	c.isRunning = false

	return nil
}

func (s FileWatch) validate() error {
	if strings.TrimSpace(s.Directory) == "" {
		return gserrors.ErrBadParam{Parameter: "Directory", Reason: "Missing value"}
	} else if strings.TrimSpace(s.DataEndpoint) == "" {
		return gserrors.ErrBadParam{Parameter: "DataEndPoint", Reason: "Missing"}
	}

	return nil
}

// Validate and prepare a SourceConfig to file pickup
func (f FileWatch) Init() error {
	if err := f.validate(); err != nil {
		return err
	}

	if _, err := os.Stat(f.Directory); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			//If no matching directory exists, create it
			err = os.MkdirAll(f.Directory, os.ModePerm)

			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	return nil
}

func (f FileWatch) IsReady() (bool, error) {
	return f.isRunning, nil
}
