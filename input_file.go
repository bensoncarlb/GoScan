package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/fsnotify/fsnotify"
)

// Steps:
// 1) create the main.go file.
// This is a special package (package main) that gets executed when you call `go run .` || `go build && ./watcher`
// 2) create the module `go mod init <some module name>`
// Don't worry much about this, just know that a module is basically a parent package.
// I ran `go mod init watcher`, which is where the watcher binary name comes from.
// 3) get the external packages you want to use
// I am using the fsnotify package (which I thought had been subsumed into the stdlib, but apparently not yet?)
// Go's package manager is crazy simple. You can just call `go get <>`.
// In this case, I ran `go get github.com/fsnotify/fsnotify`.
// 4) the actual code in the main function
// again, this is a special function and it is what go uses as the entry point for the machine code when the binary is executed.
// Go natively compiles down to machine code on all platforms.
// No need for a `mono` package or JVM or anything.

func main() {
	// based on https:github.com/fsnotify/fsnotify?tab=readme-ov-file#usage

	// spawn a watcher
	// remember that errors are a first class type in Golang
	// so checking after every call is idiomatic
	w, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}

	defer w.Close() // defers are a LIFO queue executed as the *current* function returns.
	// if you wanted to catch the error that w.Close() returns, you would do this instead:
	/*defer func() {
	  if err := w.Close(); err != nil {
	  log.Printf("failed to close watcher: %v", err)
	  }
	  }()*/

	// we are all set up, so we can now sleep the main thread.
	// for testing purposes, we could use time.Sleep().
	// time.Sleep(15 * time.Second)

	// for a more robust mechanism, we can listen for SIGINTs
	kill := make(chan os.Signal, 1) // the 1 adds a buffer of 1 item to the channel. Don't worry about that rn.
	signal.Notify(kill, os.Interrupt)

	// .Notify above will send SIGINTs on the kill channel, which we will not watch for
	//log.Println("main thread waiting on sigint...")
	//<-kill
	//log.Println("signal received. Dying...")

	// add a path for the watcher to... watch
	if err = w.Add("."); err != nil {
		panic(err)
	}
	w.Add(("C:\\Users\\Carl\\Downloads\\test"))
	for { // for loops with no conditions can only be broken out of with a `return` or a `break`

		// select is a bit more complicated as it requires understanding goroutines and channels better, but here is an explanation with all nuance striped out:
		// select is like a switch statement, but instead of "switching" on a value of a variable,
		// it "selects" a branch to execute depending on which channel (which are basically shared memory buffer? kind of? They are how goroutines communicate) has a value ready.
		// So if a watch.Event comes in, we execute that this loop.
		// If a watch.Error comes in, we execute that this loop instead.
		// That is why they are typically wrapped in infinite for loops, as it can only ever "select" a single value each loop, even if multiple are ready.
		// If both an event and an error are ready, then the selection process is arbitrary.
		select {
		case event, ok := <-w.Events:
			if !ok {
				return
			}
			
			log.Println("event:", event)
			if event.Has(fsnotify.Write) {
				log.Println("modified file:", event.Name)
			}
		case err, ok := <-w.Errors:
			if !ok {
				return
			}
			log.Println("error:", err)
		case <-kill:
			return
		}
	}
}
