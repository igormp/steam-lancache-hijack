package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync/atomic"
	"time"
)

var (
	cdnURLs = []string{
		"steampipe.akamaized.net",
		"akamai.cdn.steampipe.steamcontent.com",
		"google.cdn.steampipe.steamcontent.com",
		"level3.cdn.steampipe.steamcontent.com",
		"f3b7q2p3.ssl.hwcdn.net",
	}
	activeConnections int32
	logChannel        = make(chan string, 100000)
)

func asyncLogger() {
	for logMessage := range logChannel {
		log.Println(logMessage)
	}
}

func getRandomURL() string {
	return cdnURLs[rand.Intn(len(cdnURLs))]
}

func main() {
	go asyncLogger() // Start the async logger

	http.Handle("/depot/", http.StripPrefix("/depot/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&activeConnections, 1)
		defer atomic.AddInt32(&activeConnections, -1)

		randomURL := getRandomURL()
		redirectURL := "http://" + randomURL + "/depot/" + r.URL.Path
		logChannel <- fmt.Sprintf("Redirecting request from %s to %s", r.URL.Path, redirectURL)
		http.Redirect(w, r, redirectURL, http.StatusFound)
	})))

	go func() {
		for {
			logChannel <- fmt.Sprintf("Current active connections: %d", atomic.LoadInt32(&activeConnections))
			time.Sleep(5 * time.Second)
		}
	}()

	logChannel <- "Server starting on port 80"
	if err := http.ListenAndServe(":80", nil); err != nil {
		logChannel <- fmt.Sprintf("Error starting server: %v", err)
		close(logChannel)
	}
}
