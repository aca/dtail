package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"

	cache "github.com/Code-Hex/go-generics-cache"
	"github.com/fatih/color"
	"github.com/fsnotify/fsnotify"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	files := cache.New[string, struct{}]()
	files_lock := sync.Mutex{}

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				if event.Has(fsnotify.Write) {
					// log.Println(event.Name, event.Op)
					files_lock.Lock()
					if !files.Contains(event.Name) {
						files.Set(event.Name, struct{}{})
						go tail(event.Name)
					}
					files_lock.Unlock()
				}

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Fatal(err)
			}
		}
	}()

	err = watcher.Add(".")
	if err != nil {
		log.Fatal(err)
	}

	select {}
}

func tail(fname string) {
	fname = strings.TrimPrefix(fname, "./")
	cmd := exec.Command("tail", "-F", fname)

	stdoutPipe, _ := cmd.StdoutPipe()


	go func() {
		prefix := fmt.Sprintf("%v|", fname)
		prefix = color.RedString(prefix)

		logger := log.New(os.Stdout, prefix, 0)
		scanner := bufio.NewScanner(stdoutPipe)
		for scanner.Scan() {
			logger.Println(scanner.Text())
		}

		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}
	}()

	err := cmd.Run()
	if err != nil {
		log.Fatal(fmt.Errorf("%v: %v", fname, err))
	}
}
