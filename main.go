package main

import (
  "fmt"
	"log"
  "io/ioutil"
  "strings"
	"html/template"
	"net/http"
  "github.com/google/go-github/github"
)

// According to the test instructions, format messages like so:
//   Issue: {{Issue.title}} {{issue.url}}
//   [Opened | Closed] by {{user.login}} in {{repository.name}}
const smsTemplate = `Issue {{.Title}} {{.Url}}
{{.Action}} by {{.Username}} in {{.Repo}}
`
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
    fmt.Println("Issue event action: %s", *e.Action)
    templateData := map[string]interface{}{
      "Title":    *e.Issue.Title,
      "Url":      *e.Issue.URL,
      "Action":   *e.Action,
      "Username": *e.Sender.Login,
      "Repo":     *e.Repo.Name,
    }
    t := template.Must(template.New("sms").Parse(smsTemplate))
    builder := &strings.Builder{}
    if err := t.Execute(builder, templateData); err != nil {
      panic(err)
    }
    templatedMessageStr := builder.String()
    fmt.Println(templatedMessageStr)
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
