package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	"gopkg.in/fsnotify.v1"
)

var watcher *fsnotify.Watcher

func main() {
	sentryDSNEnv, hasDSN := os.LookupEnv("SENTRY_DSN")
	logPathEnv, hasLogPath := os.LookupEnv("LOG_PATH")
	logRawEnv, hasLogRaw := os.LookupEnv("LOG_RAW")

	isRaw := false
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

	err := sentry.Init(sentry.ClientOptions{
		Dsn: sentryDSNEnv,
	})

	if err != nil {
		log.Fatalf("sentry.Init: %s", err)
	}
	// Flush buffered events before the program terminates.
	defer sentry.Flush(2 * time.Second)

	// creates a new file watcher
	watcher, _ = fsnotify.NewWatcher()
	defer watcher.Close()

	// starting at the root of the project, walk each file/directory searching for
	// directories
	if err := filepath.Walk(logPath, watchDir); err != nil {
		fmt.Println("ERROR", err)
	}

	done := make(chan bool)

	// start parser
	go parseLog(isRaw)

	<-done
}

// watchDir gets run as a walk func, searching for directories to add watchers to
func watchDir(path string, fi os.FileInfo, err error) error {

	// since fsnotify can watch all the files in a directory, watchers only need
	// to be added to each nested directory
	if fi.Mode().IsDir() {
		return watcher.Add(path)
	}

	return nil
}
