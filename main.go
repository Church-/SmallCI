package main

import (
	"fmt"

//	"io/ioutil"
	"encoding/gob"
	"log"

	"net"
	"net/http"

//	"os"
//	"os/exec"
	
	"github.com/google/go-github/v21/github"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"

//	"gopkg.in/yaml.v2"
)

const (
	path = "/webhooks"
)


type RepoInfo struct {
	Name string
	Commit string
}

type JobResult struct {
	Status int
	Logs string
}

type MachineConfig {
	IP string
	Mem float64 
	CPU float64
}

func encodeBuffer(re RepoInfo) GobEncoder {
	var b bytes.Buffer
	gobEncoded := gob.NewEncoder(&b)
	gobEncoded.Encode(re)
	return gobEncoded
}

func handleWebhook(w http.ResponseWriter, r *http.Request, buildQueue chan, runningBuildQueue chan) {
	payload, err := github.ValidatePayload(r, []byte("Secret1"))
	//	payload, err := ioutil.ReadAll(r.Body)

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
		var repo RepoInfo{}
		repo.Name := *e.Repo.Name
		repo.Id := *e.Head
		repoCopy
		buildQueue <- repo
		runningBuildQueue <- repoCopy
		
	default:
		log.Print("Event is not among the list being acted upon.")
	}
}

func main() {
	buildQueue := make(chan RepoInfo 300)
	runningBuildQueue := make(chan RepoInfo 300)
	
	http.HandleFunc(path, handleWebhook)
	http.ListenAndServe(":3000", nil)
	server, err := net.Listen("tcp",":4545")
	conn, err := server.Accept()
	for {
		buffer,err := encodeBuffer(<- buildQueue)
		if _, err := conn.Write(buffer.Bytes()); err != nil {
			log.Print(err)
		}
		tmpBuff := make(byte[] 1000)
		if _,err := conn.Read(tmpBuff); err != nil {
			log.Print(err)
		}
		
	} 
}