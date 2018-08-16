package main

import (
  "os"
  "fmt"
	"log"
  "io/ioutil"
  "strings"
	"html/template"
	"net/http"
	"github.com/sfreiberg/gotwilio"
  "github.com/google/go-github/github"
)

// According to the test instructions, format messages like so:
//   Issue: {{Issue.title}} {{issue.url}}
//   [Opened | Closed] by {{user.login}} in {{repository.name}}
//
// This is a simple template that we'll use to fill in Issue deets:
const smsTemplate = `Issue {{.Title}} {{.Url}}
{{.Action}} by {{.Username}} in {{.Repo}}
`

// Handle requests made to our webhook
func handleWebhook(w http.ResponseWriter, r *http.Request) {

  smsNumber, ok := os.LookupEnv("TWILIO_SMS_NUMBER")
  if !ok {
    log.Fatal("Missing required environment variable TWILIO_SMS_NUMBER: set this to the recipient verified mobile number!")
  }

  // Twilio integration - account keys
  // Note: these are my test credentials, for my account jacqui.maher@gmail.com
  // https://www.twilio.com/console/project/settings
  // twilioAccountSid := "AC4e4f8c6c42a699f065badd25f5137a00"
  // twilioAuthToken := "c7ff639a05b9e03b57222a3b212364d8"
  // TODO: for a real app I wouldn't hardcode any of this, refactor if actually using...
  twilioAccountSid := "AC6f89947b675cb7faff5ca54001a888fb"
  twilioAuthToken := "19b3029e50a1dcdba16119fa8c69ef73"

  // Create a new client able to talk to Twilio's API
	twilio := gotwilio.NewTwilioClient(twilioAccountSid, twilioAuthToken)

  // Read in the request body and make sure we can parse it first:
	payload, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println("Unable to read payload body: ", err)
		return
	}
	defer r.Body.Close()

	event, err := github.ParseWebHook(github.WebHookType(r), payload)
	if err != nil {
		fmt.Println("Unable to parse webhook payload: ", err)
		return
	}

  // Although I setup this webhook to only fire on Issue-related events,
  // we still check that it's a github.IssuesEvent type
  switch e := event.(type) {
  case *github.IssuesEvent:
    fmt.Println("Received issues event from GitHub with action ", *e.Action)

    // Populate data to fill in that sms template created at the beginning:
    templateData := map[string]interface{}{
      "Title":    *e.Issue.Title,
      "Url":      *e.Issue.URL,
      "Action":   *e.Action,
      "Username": *e.Sender.Login,
      "Repo":     *e.Repo.Name,
    }

    // parse the data into the template string:
    t := template.Must(template.New("sms").Parse(smsTemplate))
    builder := &strings.Builder{}
    if err := t.Execute(builder, templateData); err != nil {
      panic(err)
    }
    templatedMessageStr := builder.String()

    // Set the sent-from number to the number I setup in the Twilio dashboard
    twilioFrom := "+61488811670"

    // Finally, send the message using the go twilio pkg
    twilio.SendSMS(twilioFrom, smsNumber, templatedMessageStr, "", "")
    fmt.Println("Sent SMS to ", smsNumber)

	default:
		log.Printf("Error, unknown event type %s\n", github.WebHookType(r))
		return
	}
}

func main() {

  // Start the http service that will accept requests from github
  log.Println("Server starting on port 8000...")
  http.HandleFunc("/webhook", handleWebhook)
  log.Fatal(http.ListenAndServe(":8000", nil))
}

