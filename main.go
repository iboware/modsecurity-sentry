package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/iboware/modsecurity-sentry/svc"
	"github.com/radovskyb/watcher"
)

func main() {
	sentryDSNEnv, hasDSN := os.LookupEnv("SENTRY_DSN")
	logPathEnv, hasLogPath := os.LookupEnv("LOG_PATH")
	logRawEnv, hasLogRaw := os.LookupEnv("LOG_RAW")
	debugEnv, hasDebugEnv := os.LookupEnv("DEBUG")
	isRaw := false
	debug := false
	logPath := "/var/log/audit"

	if !hasDSN {
		fmt.Println("SENTRY_DSN is not present")
	}

	if hasLogPath {

		if _, err := os.Stat(logPathEnv); os.IsNotExist(err) {
			log.Fatalf("Log directory is not valid: %s", err)
		}

		logPath = logPathEnv
	}

	if hasLogRaw {
		if strings.ToLower(logRawEnv) == "true" {
			isRaw = true
		}
	}

	if hasDebugEnv {
		if strings.ToLower(debugEnv) == "true" {
			debug = true
		}
	}

	err := sentry.Init(sentry.ClientOptions{
		Dsn: sentryDSNEnv,
	})

	if err != nil {
		log.Fatalf("sentry.Init: %s", err)
	} else if debug {
		fmt.Println("Logger initialized successfully.")
	}

	// Flush buffered events before the program terminates.
	defer sentry.Flush(2 * time.Second)

	w := watcher.New()

	// Only notify rename and move events.
	w.FilterOps(watcher.Write, watcher.Create)
	// Watch test_folder recursively for changes.
	if err := w.AddRecursive(logPath); err != nil {
		log.Fatalln(err)
	}
	// start parser
	go svc.WatchEvents(w, isRaw, debug)

	// Start the watching process - it'll check for changes every 100ms.
	if err := w.Start(time.Millisecond * 100); err != nil {
		log.Fatalln(err)
	}
}
