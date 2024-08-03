package core

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
)


type MessageDetails struct {
	Method string
	URL string
	Host string
	ContentLength int64
	Referer string
	Body string
}

func MessageServer() http.HandlerFunc {
	return messageHandler
}


func messageHandler(w http.ResponseWriter, r *http.Request) {
	
	var body []byte
	if r.ContentLength > 0 {
		body := make([]byte, r.ContentLength)
		r.Body.Read(body)
	}
	
	details := MessageDetails{
		Method:         r.Method,
		URL:            r.URL.String(),
		Host:           r.Host,
		ContentLength:  r.ContentLength,
		Referer:        r.Header.Get("Referer"),
		Body:           string(body),
	}
	outputBuffer, err := createNewTemplate(details)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprint(w, outputBuffer.String())
}

func createNewTemplate(messageDetails MessageDetails) (bytes.Buffer , error){
	const messageTemplate = `
	Method: {{.Method}}
	URL: {{.URL}}
	HTTP Version: {{.HTTPVersion}}
	Host: {{.Host}}
	Content-Length: {{.ContentLength}}
	Referer: {{.Referer}}
	Body: {{.Body}}
	`

	tmpl, err := template.New("messageDetails").Parse(messageTemplate)

	if err != nil {
		return *bytes.NewBuffer([]byte{}), err
	}
	

	var outputBuffer bytes.Buffer
	
	err = tmpl.Execute(&outputBuffer, messageDetails)
	if err != nil {
		return *bytes.NewBuffer([]byte{}), err
	}

	return outputBuffer, nil

}
