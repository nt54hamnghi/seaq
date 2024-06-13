package main

import (
	"log"
	"sync"
	"time"

	dumplingai "github.com/nt54hamnghi/hoc/dumpling-ai"
	"github.com/nt54hamnghi/hoc/util"
)

type task struct {
	filename string
	genFunc  func(string) (string, error)
}

func (t task) spawn(data string) error {

	msg, err := t.genFunc(data)
	if err != nil {
		return err
	}

	if err := util.WriteFile(t.filename, msg); err != nil {
		return err

	}
	return nil
}

func main() {
	log.SetFlags(log.Ltime)
	errChan := make(chan error, 2)

	transcript, err := dumplingai.GetTranscript("https://www.youtube.com/watch?v=a620mE9nyBQ")
	if err != nil {
		log.Fatal("Failed to get transcript: ", err)
	}

	tasks := []task{
		// {
		// 	filename: "summary.md",
		// 	genFunc:  openai.Prime,
		// },
		// {
		// 	filename: "connection.md",
		// 	genFunc:  openai.MakeConnection,
		// },
		{
			filename: "summary.md",
			genFunc: func(s string) (string, error) {
				time.Sleep(3 * time.Second)
				return "summary", nil
			},
		},
		{
			filename: "connection.md",
			genFunc: func(s string) (string, error) {
				time.Sleep(3 * time.Second)
				return "connection", nil
			},
		},
	}

	performTasks(tasks, errChan, transcript)

	for e := range errChan {
		if e != nil {
			log.Fatal(e)
		}
	}
}

func performTasks(tasks []task, errChan chan<- error, data string) {
	wg := sync.WaitGroup{}
	for _, t := range tasks {
		wg.Add(1)

		go func(t task) {
			defer wg.Done()

			log.Println("Processing", t.filename)
			errChan <- t.spawn(data)
			log.Println("Done with", t.filename)
		}(t)
	}

	go func() {
		wg.Wait()
		close(errChan)
	}()
}
