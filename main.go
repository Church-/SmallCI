package main

import (
	"fmt"

	"encoding/gob"
	"log"
	"sync"
	
	"net"
	"net/http"

	
	"github.com/google/go-github/v21/github"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"

)

const (
	path = "/webhooks"
)


type Job struct {
	Name string
	Commit string
}

type JobResult struct {
	Status int
	Logs string
}

type Worker struct {
	IP string
	Mem float64 
	CPU float64
	connMu sync.Mutex; 
	conn net.Conn 
}

func (w *Worker) SendJob(j *Job) error { 
	w.connMu.Lock()
	defer w.connMu.Unlock()
	_, err := w.conn.Write()
	return err 
}


func encodeBuffer(job Job) GobEncoder {
	var b bytes.Buffer
	gobEncoded := gob.NewEncoder(&b)
	gobEncoded.Encode(job)
	return gobEncoded
}

func handleWebhook(w http.ResponseWriter, r *http.Request, buildQueue chan, runningBuildQueue chan) {
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
		var job Job{}
		job.Name := *e.Repo.Name
		job.Id := *e.Head
		jobCopy := job
		buildQueue <- job
		runningBuildQueue <- repoCopy
		
	default:
		log.Print("Event is not among the list being acted upon.")
	}
}

func handleTcpConnection(c net.Conn) {
	var worker Worker{}
	
	buffer, err := encodeBuffer(<- buildQueue)
	if _, err := conn.Write(buffer.Bytes()); err != nil {
		log.Print(err)
	}
	tmpBuff := make(byte[] 1000)
	if _, err := conn.Read(tmpBuff); err != nil {
		log.Print(err)
	}
	
}

func main() {
	workerGroup := make(chan *Worker 30)
	buildQueue := make(chan Job 300)
	runningBuildQueue := make(chan Job 300)
	finishedQueue := make(chan JobResult 300)
	
	http.HandleFunc(path, handleWebhook)
	http.ListenAndServe(":3000", nil)
	
	server, err := net.Listen("tcp",":4545")

	for {
		if conn, err := server.Accept(); err != nil {
			log.Print(err)
		}
		go handleTcpConnection(conn)
	} 
}