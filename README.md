# ght

A GitHub-Twilio integration that notifies via SMS when an Issue is opened or closed on this repo.

## Setup

This version of the app only runs locally and requires you to have `ngrok` setup. Please make sure you install it by following the instructions for your OS at [https://ngrok.com/download](https://ngrok.com/download) before proceeding.

First clone this repo into your $GOPATH:

```
mkdir -p $GOPATH/github/jacqui
cd $GOPATH/github/jacqui
git clone git@github.com:jacqui/ght.git
```

Ensure you have a verified mobile number in Twilio - do this in your Twilio dashboard]().

Before running this service, set up that number in your environment:

```
export TWILIO_SMS_NUMBER="+61555555555"
```

Then run the app:

```
cd ght
go run main.go
```

In a terminal window, start up ngrok and run on port 8000:

```
ngrok http 8000
```

When it outputs the URLs it's mapping your local server to, copy the https version. We'll use that to setup the webhook on this github repo:

* open [https://github.com/jacqui/ght/settings/hooks/new](https://github.com/jacqui/ght/settings/hooks/new)
* enter your local webhook endpoint in the `Payload URL` field, for example: "https://29341861.ngrok.io/webhook"
* choose content type "application/json"
* leave the secret blank for this exercise, and keep ssl verification enabled
* under "Which events would you like to trigger this webhook?" choose "Let me select individual events"
* check the box next to "Issues" and uncheck the box next to "Pushes" - the only selection should be "Issues"
* ensure the box next to "Active" at the very end is checked
* click the "Add webhook" button

## Running

You can test that all integration points are working by opening or closing an issue on this github repo: 

* [open a new issue](https://github.com/jacqui/ght/issues/new)
* [close this sample issue](https://github.com/jacqui/ght/issues/1)

## Note

This very basic GitHub/Twilio integration service will only run locally, requires ngrok, and will only SMS my own mobile number using a test Twilio sender number in Australia. It was created as part of a job interview, so please, don't use this in production anywhere!

