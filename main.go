package main

import (
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/google/go-github/github"
	"github.com/sfreiberg/gotwilio"
)

// According to the test instructions, format messages like so:
//   Issue: {{Issue.title}} {{issue.url}}
//   [Opened | Closed] by {{user.login}} in {{repository.name}}
//
// This is a simple template that we'll use to fill in Issue deets:
const smsTemplate = `Issue {{.Title}} {{.Url}}
{{.Action}} by {{.Username}} in {{.Repo}}
`

type twilioClient interface {
	SendSMS(from, to, body, statusCallback, applicationSid string) (smsResponse *gotwilio.SmsResponse, exception *gotwilio.Exception, err error)
}

func newTwilioClient() (twilioClient, error) {
	accountSID, ok := os.LookupEnv("TWILIO_ACCOUNT_SID")
	if !ok {
		return nil, errors.New("Missing required environment variable TWILIO_ACCOUNT_SID")
	}
	authToken, ok := os.LookupEnv("TWILIO_AUTHTOKEN")
	if !ok {
		return nil, errors.New("Missing required environment variable TWILIO_AUTHTOKEN")
	}
	return gotwilio.NewTwilioClient(accountSID, authToken), nil
}

type webhookEventParseFunc func(r *http.Request, payload []byte) (*github.IssuesEvent, error)

var parseWebhookEvent = func(r *http.Request, payload []byte) (*github.IssuesEvent, error) {
	whType := github.WebHookType(r)
	event, err := github.ParseWebHook(whType, payload)
	if err != nil {
		return nil, err
	}
	issuesEvent, ok := event.(*github.IssuesEvent)
	if !ok {
		return nil, fmt.Errorf("unknown event type %s", whType)
	}
	return issuesEvent, nil
}

type smsBuilder func(e *github.IssuesEvent, smsTemplate string) (string, error)

var buildTemplate = func(e *github.IssuesEvent, smsTemplate string) (string, error) {
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
		return "", err
	}
	templatedMessageStr := builder.String()
	return templatedMessageStr, nil
}

func twillioHandler(client twilioClient, parseWhEvent webhookEventParseFunc, sb smsBuilder) http.HandlerFunc {
	smsNumber, ok := os.LookupEnv("TWILIO_SMS_NUMBER")
	if !ok {
		log.Fatal("Missing required environment variable TWILIO_SMS_NUMBER: set this to the recipient verified mobile number!")
	}
	return func(w http.ResponseWriter, r *http.Request) {
		payload, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Println("Unable to read payload body: ", err)
			return
		}
		defer r.Body.Close()

		e, err := parseWhEvent(r, payload)
		if err != nil {
			log.Println("Unable to parse webhook payload for IssuesEvent: ", err)
			return
		}

		fmt.Println("Received issues event from GitHub with action ", *e.Action)

		// Set the sent-from number to the number I setup in the Twilio dashboard
		twilioFrom := "+61488811670"
		sms, err := sb(e, smsTemplate)
		if err != nil {
			log.Println("couldn't build sms message:", err)
			return
		}

		// Finally, send the message using the go twilio pkg
		client.SendSMS(twilioFrom, smsNumber, sms, "", "")
		log.Println("Sent SMS to ", smsNumber)

	}
}

func main() {

	tw, err := newTwilioClient()
	if err != nil {
		log.Fatal(err)
	}
	// Start the http service that will accept requests from github
	log.Println("Server starting on port 8000...")
	http.HandleFunc("/webhook", twillioHandler(tw, parseWebhookEvent, buildTemplate))
	log.Fatal(http.ListenAndServe(":8000", nil))
}
