package main

import (
  "fmt"
	"io"
	"log"
	"net/http"
	"os"
)

// Handle requests made to our webhook
func handleWebhook(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("headers: %v\n", r.Header)

	_, err := io.Copy(os.Stdout, r.Body)
	if err != nil {
		log.Println(err)
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
