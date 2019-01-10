package main

import (
	"encoding/gob"
	"bytes"
	"log"
	"sync"

	"net"
	"net/http"

	"github.com/google/go-github/v21/github"
)

const (
	path = "/webhooks"
)

type Job struct {
	Name   string
	URL    string
	Commit string
}

type JobResult struct {
	Status int
	Logs   string
}

type Worker struct {
	Mem    float64
	CPU    float64
	connMu *sync.Mutex
	conn   net.Conn
}

func (w *Worker) SendJob(j Job) error {
	w.connMu.Lock()
	defer w.connMu.Unlock()
	buffer := encodeBuffer(j)
	_, err := w.conn.Write(buffer.Bytes())
	if err != nil {
		log.Print(err)
	}
	return err
}

func encodeBuffer(job Job) bytes.Buffer {
	b := new(bytes.Buffer)
	gobEncoded := gob.NewEncoder(b)
	err := gobEncoded.Encode(job)
	if err != nil {
		log.Print(err)
	}
	return *b
}

func handleWebhook(w http.ResponseWriter, r *http.Request) {
	payload, err := github.ValidatePayload(r, []byte("Secret1"))

	if err != nil {
		log.Printf("Error validating payload: err=%v", err)
		return
	}

	defer r.Body.Close()

	event, err := github.ParseWebHook(github.WebHookType(r), payload)
	if err != nil {
		log.Printf("Couldn't parse webhook: err=%v", err)
		return
	}

	switch e := event.(type) {

	case *github.PushEvent:
		job := Job{Name: *e.Repo.Name, Commit: *e.Head, URL: *e.Repo.GitURL}
		
	default:
		log.Print("Event is not among the list being acted upon.")
	}
}

func handleTcpConnection(w *Worker) {
	tmpBuff := make([]byte, 1000)
	if _, err := w.conn.Read(tmpBuff); err != nil {
		log.Print(err)
	}
	
}

func main() {
	workerGroup := make([]*Worker, 0)
	buildQueue := make([]Job, 0)
	runningBuildQueue := make([]Job, 0)
	finishedQueue := make(chan JobResult, 300)

	http.HandleFunc(path, handleWebhook)
	http.ListenAndServe(":3000", nil)

	server, err := net.Listen("tcp", ":4545")

	for {
		conn, err := server.Accept()
		if err != nil {
			log.Print(err)
		}
		w := Worker{conn: conn}
		copy(workerGroup, []*Worker{&w})
		go handleTcpConnection(&w)
		if buildQueue.isEmpty() { //Need to fix only the if statement, .isEmpty is not a real function for channels.
			for _, elem := range workerGroup {
				if elem.Mem <= 50 && elem.CPU <= 50 {
					elem.SendJob(<-buildQueue)
				}
			}
		}
	}
}
