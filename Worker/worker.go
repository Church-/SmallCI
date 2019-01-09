package main

import (
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

	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/cpu"
)

type Pipeline struct {
	Kind  string
	Steps []struct {
		Name     string
		Commands []string
	}
}

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

func runPipeline(p *Pipeline) {
	if p.Kind != "pipeline" {
		log.Print("Not a pipeline.")
	}
	for _, step := range p.Steps {
		fmt.Printf("Step: %s has started", step.Name)
		for _, command := range step.Commands {
			cmd := exec.Command(command)
			fmt.Printf("Command: %s is running.", command)
			if err := cmd.Run(); err != nil {
				log.Printf("Command finished with error: %v", err)
			}
		}
	}
}




func handleConnection(j Job) {
	path := "/tmp/" + *j.Name
	if err := os.Chdir(path); err != nil {
		os.Mkdir(path, 0777)
		_, err := git.PlainClone(path, false, &git.CloneOptions{
			URL: j.GitURL,
		})
		if err != nil {
			log.Print(err)
		}
	}else {
			os.Chdir(path)
			req, _ := git.PlainOpen(path)
			w, _ := req.Worktree()
			commit := plumbing.NewHash(j.ID)
			w.Reset(&git.ResetOptions{
				Mode:   git.HardReset,
				Commit: commit})
	}
	os.Chdir(path)
	go runPipeline(parseYamlPipeLine())
}

func main() {
	if conn, err := net.Dial("tcp","MASTER_ADDR"); err != nil {
		log.Print(err)
	}

	for {
		
	}
}
