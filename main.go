package main

import (
	"fmt"

	"io/ioutil"

	"log"

	"net/http"

	"os"
	"os/exec"
	
	"github.com/google/go-github/v21/github"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"

	"gopkg.in/yaml.v2"
)

const (
	path = "/webhooks"
)

type Pipeline struct {
	Kind  string
	Steps []struct {
		Name     string
		Commands []string
	}
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

func handleWebhook(w http.ResponseWriter, r *http.Request) {
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
		path := "/tmp/" + *e.Repo.Name
		if err := os.Chdir(path); err != nil {
			os.Mkdir(path, 0777)
			_, err := git.PlainClone(path, false, &git.CloneOptions{
				URL: *e.Repo.GitURL,
			})
			if err != nil {
				log.Print(err)
			}
		} else {
			os.Chdir(path)
			req, _ := git.PlainOpen(path)
			w, _ := req.Worktree()
			commit := plumbing.NewHash(*e.HeadCommit.ID)
			w.Reset(&git.ResetOptions{
				Mode:   git.HardReset,
				Commit: commit})
		}
		os.Chdir(path)
		go runPipeline(parseYamlPipeLine())

	default:
		log.Print("Event is not among the list being acted upon.")
	}
}

func main() {
	http.HandleFunc(path, handleWebhook)
	http.ListenAndServe(":3000", nil)
}
