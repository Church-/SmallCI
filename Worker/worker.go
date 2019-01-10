package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"

	"os"
	"os/exec"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"

	"gopkg.in/yaml.v2"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

type Pipeline struct {
	Kind  string
	Steps []struct {
		Name     string
		Commands []string
	}
}

type Job struct {
	Name   string
	Commit string
	URL    string
}

type JobResult struct {
	Status int
	Logs   string
}

type Worker struct {
	Mem float64
	CPU float64
}

func encodeBuffer(obj interface{}) bytes.Buffer {
	b := new(bytes.Buffer)
	gobEncoded := gob.NewEncoder(b)
	switch obj.(type) {
		case Worker:
		
		case JobResult:

	}
	err := gobEncoded.Encode(obj)
	if err != nil {
		log.Print(err)
	}
	return *b
}

func parseYamlPipeLine() *Pipeline {
	var p *Pipeline
	file, err := ioutil.ReadFile("./config.yaml")
	if err != nil {
		log.Print(err)
	}
	if err := yaml.Unmarshal(file, p); err != nil {
		log.Print(err)
	}
	return p
}

func runPipeline(p *Pipeline) JobResult {
	jr := JobResult{}
	out := new(bytes.Buffer)
	multi := io.MultiWriter(os.Stdout, out)

	if p.Kind != "pipeline" {
		log.Print("Not a pipeline.")
		jr.Status = 1
		return jr
	}
	for _, step := range p.Steps {
		fmt.Printf("Step: %s has started", step.Name)
		for _, command := range step.Commands {
			cmd := exec.Command(command)
			fmt.Printf("Command: %s is running.", command)
			cmd.Stdout = multi
			if err := cmd.Run(); err != nil {
				log.Printf("Command finished with error: %v", err)
				jr.Status = 1
				jr.Logs = out.String()
				return jr
			}
		}
	}
	jr.Status = 1
	jr.Logs = out.String()
	return jr
}

func handleConnection(j Job) {
	path := "/tmp/" + j.Name
	if err := os.Chdir(path); err != nil {
		os.Mkdir(path, 0777)
		_, err := git.PlainClone(path, false, &git.CloneOptions{
			URL: j.URL,
		})
		if err != nil {
			log.Print(err)
		}
	} else {
		os.Chdir(path)
		req, _ := git.PlainOpen(path)
		w, _ := req.Worktree()
		commit := plumbing.NewHash(j.Commit)
		w.Reset(&git.ResetOptions{
			Mode:   git.HardReset,
			Commit: commit})
	}
	os.Chdir(path)
	go runPipeline(parseYamlPipeLine())
}

func main() {
	resultQueue := make(chan JobResult, 10)
	conn, err := net.Dial("tcp", "localhost:4545")
	if err != nil {
		log.Print(err)
	}
	w := Worker{}
	v, _ := mem.VirtualMemory()
	for {
		u, _ := cpu.Percent(0, false)
		w.Mem = v.UsedPercent
		w.CPU = u[0]

		n, err := conn.Write(encodeBuffer(w))
		if err != nil {
			log.Print(err)
		}
		buffer,err := ioutil.ReadAll(conn)
		if err != nil {
			log.Print(err)
		}
	}
}