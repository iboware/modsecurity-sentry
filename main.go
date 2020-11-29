package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/iboware/modsecurity-sentry/svc"
	"gopkg.in/fsnotify.v1"
)

var watcher *fsnotify.Watcher

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

	// creates a new file watcher
	watcher, _ = fsnotify.NewWatcher()
	defer watcher.Close()

	// starting at the root of the project, walk each file/directory searching for
	// directories
	go watchPeriodically(logPath, 5)
	done := make(chan bool)

	// start parser
	go svc.WatchEvents(watcher, isRaw, debug)

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

// watchPeriodically triggers watchDir function periodically.
func watchPeriodically(directory string, interval int) {
	done := make(chan struct{})
	go func() {
		done <- struct{}{}
	}()
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()
	for ; ; <-ticker.C {
		<-done
		if err := filepath.Walk(directory, watchDir); err != nil {
			fmt.Println(err)
		}
		go func() {
			done <- struct{}{}
		}()
	}
}
