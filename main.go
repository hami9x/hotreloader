package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
	flags "github.com/jessevdk/go-flags"

	"github.com/howeyc/fsnotify"
)

var opts struct {
	Program string `short:"p" long:"program" description:"The program to run for hot code reload" required:"true"`
	Args    string `short:"a" long:"args" description:"Arguments to pass to the program" optional:"true"`
	Dir     string `short:"d" long:"dir" description:"Directory to run the program in" optional:"true"`
}

func main() {
	_, err := flags.Parse(&opts)
	if err != nil {
		fmt.Printf(err.Error())
		os.Exit(1)
	}
	exe, cargs, dir := opts.Program, opts.Args, opts.Dir
	if dir != "" {
		err := os.Chdir(dir)
		if err != nil {
			panic(err.Error())
		}
	}
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	done := make(chan bool)

	// Process events
	go func() {
		for {
			select {
			case ev := <-watcher.Event:
				if ev.IsModify() && filepath.Ext(ev.Name) == ".go" {
					time.Sleep(time.Second)
					for nev, cont := <-watcher.Event; cont; {
						ev = nev
						select {
						case nev = <-watcher.Event:
							continue
						default:
							cont = false
						}
					}
					log.Println("event:", ev)
					log.Printf("%s", exe)
					cmd := exec.Command(exe, strings.Split(cargs, " ")...)
					cmd.Stdout = os.Stdout
					cmd.Stderr = os.Stderr
					cmd.Run()
					if err != nil {
						log.Println("%s", err.Error())
					}
				}
			case err := <-watcher.Error:
				panic(err.Error())
				return
			}
		}
	}()

	cwd, _ := os.Getwd()
	err = watcher.Watch(cwd)
	if err != nil {
		panic(err.Error())
	}

	<-done

	/* ... do stuff ... */
	watcher.Close()
}
