package svc

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/getsentry/sentry-go"
	"github.com/iboware/modsecurity-sentry/model"
	"github.com/radovskyb/watcher"
)

//watchEvents watches create file events and logs them into sentry.
func WatchEvents(w *watcher.Watcher, isRaw bool, debug bool) {

	if debug {
		fmt.Println("Waiting for events...")
	}

	for {
		select {
		// watch for events
		case event := <-w.Event:
			if event.IsDir() {
				continue
			}
			if event.Op == watcher.Create || event.Op == watcher.Write {
				if debug {
					fmt.Println("Event:", event)
				}
				jsonFile, err := os.Open(event.Path)
				// if we os.Open returns an error then handle it
				if err != nil {
					fmt.Println(err)
				}
				// defer the closing of our jsonFile so that we can parse it later on
				defer jsonFile.Close()

				byteValue, err := ioutil.ReadAll(jsonFile)
				if err == nil {
					logEvent(byteValue, isRaw, debug)
				} else if debug {
					fmt.Println("ERROR:", err)
				}
			}

			// watch for errors
		case err := <-w.Error:
			fmt.Println("ERROR", err)
		}
	}
}

func logEvent(event []byte, isRaw bool, debug bool) {
	var entry model.ModsecurityLogEntry
	var sentryEvent *sentry.Event

	// we unmarshal our byteArray which contains log entry
	json.Unmarshal(event, &entry)

	sentryEvent = sentry.NewEvent()
	sentryEvent.Request = new(sentry.Request)
	sentryEvent.Request.URL = entry.Transaction.Request.URI
	sentryEvent.Request.Method = entry.Transaction.Request.Method

	sentryEvent.Request.Headers = make(map[string]string)
	sentryEvent.Request.Headers["Host"] = entry.Transaction.Request.Headers.Host
	sentryEvent.Request.Headers["Accept"] = entry.Transaction.Request.Headers.Accept
	sentryEvent.Request.Headers["UserAgent"] = entry.Transaction.Request.Headers.UserAgent
	sentryEvent.Level = sentry.LevelError

	sentryEvent.Transaction = entry.Transaction.UniqueID
	if len(entry.Transaction.Messages) > 0 {
		for _, message := range entry.Transaction.Messages {
			sentryEvent.Message = createMessage(message)
			sentryEvent.Tags = createTags(message, entry.Transaction)
		}
	} else {
		sentryEvent.Message = entry.Transaction.Request.URI
	}

	if isRaw {
		sentryEvent.Message += "Request:" + "\n"
		sentryEvent.Message += "------------" + "\n"
		prettyRequest, prettyRequestErr := json.MarshalIndent(entry.Transaction.Request, "", " ")
		if prettyRequestErr == nil {
			sentryEvent.Message += string(prettyRequest) + "\n"
		} else {
			fmt.Println("ERROR", prettyRequestErr)
		}

		sentryEvent.Message += "Response:" + "\n"
		sentryEvent.Message += "------------" + "\n"
		prettyResponse, prettyResponseErr := json.MarshalIndent(entry.Transaction.Response, "", " ")
		if prettyResponseErr == nil {
			sentryEvent.Message += string(prettyResponse) + "\n"
		} else {
			fmt.Println("ERROR", prettyResponseErr)
		}
	}

	sentryEvent.Tags["timestamp"] = entry.Transaction.TimeStamp
	sentryEvent.Tags["client_ip"] = entry.Transaction.ClientIP

	eventId := sentry.CaptureEvent(sentryEvent)
	if debug {
		fmt.Println("Captured Event: ", eventId)
	}

}

// createTags creates tags for sentry event.
func createTags(m model.Message, t model.Transaction) map[string]string {
	var tags = make(map[string]string)

	if m.Details.Accuracy != "" {
		tags["accuracy"] = m.Details.Accuracy
	}
	if m.Details.Maturity != "" {
		tags["maturity"] = m.Details.Maturity
	}
	if m.Details.Rev != "" {
		tags["rev"] = m.Details.Rev
	}
	if m.Details.RuleID != "" {
		tags["rule_id"] = m.Details.RuleID
	}
	if m.Details.Severity != "" {
		tags["severity"] = m.Details.Severity
	}
	if m.Details.Ver != "" {
		tags["ver"] = m.Details.Ver
	}
	if t.Request.Headers.Host != "" {
		tags["host"] = t.Request.Headers.Host
	}
	if t.Request.Headers.UserAgent != "" {
		tags["user_agent"] = t.Request.Headers.UserAgent
	}
	return tags
}

// createMessage creates message for sentry event.
func createMessage(m model.Message) string {
	var message string
	message += m.Message + "\n"
	message += "----------\n"
	if m.Details.Match != "" {
		message += "Match: " + m.Details.Match + "\n"
	}
	if m.Details.File != "" {

		message += "File: " + m.Details.File + "#" + m.Details.LineNumber + "\n"
	}
	if m.Details.Data != "" {
		message += "Data:" + m.Details.Data + "\n"
	}

	return message
}
