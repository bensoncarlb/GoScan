// Data Souce module from File Pickups
package sourceFile

//TODO  move to Sources type package for all sourcesa
import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/bensoncb/GoScan/internal/gsRecord"
	"github.com/bensoncb/GoScan/internal/gserrors"
	"github.com/fsnotify/fsnotify"
	"github.com/rflandau/expiring"
)

// TODO rename to a ~FileConfig
type SourceConfig struct {
	isRunning    bool
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
func fileEvents(chFiles <-chan string, DataEndpoint string) {
	for {
		file, ok := <-chFiles

		if !ok {
			return
		}
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
	}
}

/*
* Start the FileWatch
 */
func (c *SourceConfig) Start() error {
	// TODO track and check if already started
	if c.isRunning {
		return nil
	}

	var err error
	c.fsWatch, err = fsnotify.NewBufferedWatcher(30)

	if err != nil {
		return err
	}

	log.Printf("Watching %s", c.Directory)
	if err := c.fsWatch.Add(c.Directory); err != nil {
		return err
	}

	//TODO configurable
	c.chFiles = make(chan string, 50)

	go fileWatch(c.fsWatch, c.chFiles)

	//Spawn worker

	//TODO configurable
	for range 1 {
		go fileEvents(c.chFiles, c.DataEndpoint)
	}

	c.isRunning = true

	return nil
}

/*
* Cleanly stop the FileWatcher
 */
func (c *SourceConfig) Stop() error {
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

/*
* Validate and prepare a SourceConfig to file pickup
 */
func New(Directory string, DataEndpoint string) (SourceConfig, error) {
	//TODO strings.TrimSpace(Directory) = ""
	if len(Directory) == 0 {
		return SourceConfig{}, gserrors.ErrBadParam{Parameter: "Directory", Reason: "Missing value"}
	} else if len(DataEndpoint) == 0 {
		return SourceConfig{}, gserrors.ErrBadParam{Parameter: "DataEndPoint", Reason: "Missing"}
	}

	if _, err := os.Stat(Directory); errors.Is(err, os.ErrNotExist) {
		err = os.Mkdir(Directory, os.ModePerm)
		if err != nil {
			return SourceConfig{}, err
		}
	} else if err != nil {
		return SourceConfig{}, err
	}

	return SourceConfig{Directory: Directory, DataEndpoint: DataEndpoint}, nil
}
