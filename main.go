package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"runtime"
	"sync"
	"time"
)

var (
	counter int
	mu      sync.Mutex
)

// FunFact represents the structure of the fun fact fetched from the API.
type FunFact struct {
	Text string `json:"text"`
}

func main() {
	port := getPort()
	http.HandleFunc("/", helloHandler)
	http.HandleFunc("/submit", submitHandler)

	log.Printf("Listening on port %s\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}

func getPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	return port
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	counter++
	mu.Unlock()

	currentTime := time.Now().Format("Mon Jan 2 15:04:05 MST 2006")
	hostname := getHostname()
	cpuCount, memAlloc := getSystemInfo()
	funFact := getFunFact()

	response := fmt.Sprintf(
		`<html>
			<head>
				<title>Hello from CloudStation</title>
			</head>
			<body>
				<h1>Hello from CloudStation</h1>
				<p>Current time: %s</p>
				<p>Server hostname: %s</p>
				<p>Page visits: %d</p>
				<p>CPU Count: %d</p>
				<p>Memory Alloc: %d bytes</p>
				<h2>Fun Fact:</h2>
				<p>%s</p>
				<h2>Submit a Message:</h2>
				<form action="/submit" method="post">
					<input type="text" name="message" placeholder="Enter your message">
					<button type="submit">Submit</button>
				</form>
			</body>
		</html>`,
		currentTime, hostname, counter, cpuCount, memAlloc, funFact,
	)

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, response)
}

func submitHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		message := r.FormValue("message")
		response := fmt.Sprintf(
			`<html>
				<head>
					<title>Message Submitted</title>
				</head>
				<body>
					<h1>Thank you for your message!</h1>
					<p>Your message: %s</p>
					<a href="/">Go back</a>
				</body>
			</html>`,
			message,
		)

		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, response)
	} else {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return hostname
}

func getSystemInfo() (int, uint64) {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	return runtime.NumCPU(), memStats.Alloc
}

func getFunFact() string {
	resp, err := http.Get("https://uselessfacts.jsph.pl/random.json?language=en")
	if err != nil {
		return "Could not fetch a fun fact."
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "Could not read fun fact response."
	}

	var fact FunFact
	err = json.Unmarshal(body, &fact)
	if err != nil {
		return "Could not parse fun fact response."
	}

	return fact.Text
}
