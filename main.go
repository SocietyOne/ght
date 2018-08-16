package main

import (
  "fmt"
	"log"
  "io/ioutil"
	"net/http"
  "github.com/google/go-github/github"
)

// Handle requests made to our webhook
func handleWebhook(w http.ResponseWriter, r *http.Request) {

	payload, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("error reading request body: err=%s\n", err)
		return
	}
	defer r.Body.Close()

	event, err := github.ParseWebHook(github.WebHookType(r), payload)
	if err != nil {
		log.Printf("could not parse webhook: err=%s\n", err)
		return
	}
  switch e := event.(type) {
  case *github.IssuesEvent:
		// this is an issue-related event
    fmt.Println("issue type found: %s", *e.Action)
    fmt.Println("issue title: %s", *e.Issue.Title)
    fmt.Println("issue creator: %s", *e.Sender.Login)
    fmt.Println("issue repo: %s", *e.Repo.Name)
	default:
		log.Printf("Error, unknown event type %s\n", github.WebHookType(r))
		return
	}
}

func main() {
  // start the http service that will accept requests from github
  log.Println("server started")
  http.HandleFunc("/webhook", handleWebhook)
  log.Fatal(http.ListenAndServe(":8000", nil))
}

// TODO: other stuff to keep in mind:
//        * properly formatting this code
//        * what test framework to use
//        * ensure the README makes sense. see note in that file.
