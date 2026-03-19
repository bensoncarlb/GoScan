package sourceFile

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	paramerror "github.com/bensoncb/GoScan/internal/errors"
	"github.com/bensoncb/GoScan/internal/structs/inputFile"
	"github.com/fsnotify/fsnotify"
)

type SourceConfig struct {
	Directory    string
	DataEndpoint string
	fsWatch      *fsnotify.Watcher
}

func filewatcher(fsWatch *fsnotify.Watcher, DataEndpoint string) {
	log.Printf("File watcher starting up")
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

				i := inputFile.InputFile{Size: len(f), Name: filepath.Base(event.Name), Src: "file", Data: f}

				b := new(bytes.Buffer)

				err = json.NewEncoder(b).Encode(i)
				if err != nil {
					panic(err)
				}

				resp, err := http.Post(DataEndpoint, "application/json", b) //TODO Move to channel or direct?

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

	//Spawn worker
	for range 1 {
		go filewatcher(c.fsWatch, c.DataEndpoint)
	}

	return err
}

func (c *SourceConfig) Stop() error {
	c.fsWatch.Close()
	return nil
}

func New(Directory string, DataEndpoint string) (SourceConfig, error) {
	var err error

	if len(Directory) == 0 {
		err = paramerror.ErrBadParam{Parameter: "Directory", Reason: "Missing value"}
	} else if len(DataEndpoint) == 0 {
		err = paramerror.ErrBadParam{Parameter: "DataEndPoint", Reason: "Missing"}
	}

	if err == nil {
		_, errFile := os.Stat(Directory)

		if errors.Is(errFile, os.ErrNotExist) {
			errFile = os.Mkdir(Directory, os.ModePerm)

			if errFile != nil {
				err = errFile
			}
		}
	}

	conf := SourceConfig{Directory: Directory, DataEndpoint: DataEndpoint}

	return conf, err
}
