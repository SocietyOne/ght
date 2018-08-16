package main

import (
  "fmt"
	"log"
  "io/ioutil"
  "strings"
	"html/template"
  "net/url"
	"net/http"
  "encoding/json"
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

  // Twilio integration - account keys
  // Note: these are my test credentials, for my account jacqui.maher@gmail.com
  // https://www.twilio.com/console/project/settings
  // twilioAccountSid := "AC4e4f8c6c42a699f065badd25f5137a00"
  // twilioAuthToken := "c7ff639a05b9e03b57222a3b212364d8"
  twilioAccountSid := "AC6f89947b675cb7faff5ca54001a888fb"
  twilioAuthToken := "19b3029e50a1dcdba16119fa8c69ef73"

  // // Format the Twilio Messages API URL with testing account id
  twilioUrl := "https://api.twilio.com/2010-04-01/Accounts/" + twilioAccountSid + "/Messages.json"
  
  fmt.Println(twilioUrl)

  // Read in the request body and make sure we can parse it first:
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

  // Although I setup this webhook to only fire on Issue-related events,
  // we still check that it's a github.IssuesEvent type
  switch e := event.(type) {
  case *github.IssuesEvent:
    fmt.Println("Issue event fired with action: %s", *e.Action)

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
    // for now, output it to stdout
    fmt.Println(templatedMessageStr)

    twilioMessageData := url.Values{}

    // Set the destination number to my verified Twilio mobile
    twilioMessageData.Set("To","+61491081106")

    // Set the sent-from number to the number I setup in the Twilio dashboard
    twilioMessageData.Set("From","+61488811670")

    // Finally set the text to the formatted string with the GH issue content
    twilioMessageData.Set("Body", templatedMessageStr)
    twilioMessageDataReader := *strings.NewReader(twilioMessageData.Encode())
    log.Println(twilioMessageDataReader)

    // post to Twilio's API to send the sms
    client := &http.Client{}
    req, _ := http.NewRequest("POST", twilioUrl, &twilioMessageDataReader)
    req.SetBasicAuth(twilioAccountSid, twilioAuthToken)
    //req.Header.Add("Accept", "application/json")
    //req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

    // Make HTTP POST request and return the message SID
    resp, _ := client.Do(req)
    if (resp.StatusCode >= 200 && resp.StatusCode < 300) {
      var data map[string]interface{}
      decoder := json.NewDecoder(resp.Body)
      err := decoder.Decode(&data)
      if (err == nil) {
        fmt.Println(data["sid"])
      }
    } else {
      fmt.Println(resp.Status);
      fmt.Println(resp);
    }

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
