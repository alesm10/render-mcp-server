package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

type Message struct {
	Sender  string `json:"sender"`
	Message string `json:"message"`
	Time    string `json:"time"`
}

func handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	var msg Message
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	log.Printf("📩 Received message: %+v\n", msg)

	makeWebhook := os.Getenv("MAKE_WEBHOOK_URL")
	if makeWebhook == "" {
		http.Error(w, "MAKE_WEBHOOK_URL not set", http.StatusInternalServerError)
		return
	}

	payload, _ := json.Marshal(msg)
	resp, err := http.Post(makeWebhook, "application/json", 
	                      http.NoBody)
	if err != nil {
		log.Printf("❌ Error sending to Make: %v\n", err)
		http.Error(w, "Failed to forward to Make", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	log.Printf("➡️ Sent to Make: %s | Status: %s\n", makeWebhook, resp.Status)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "OK")
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	http.HandleFunc("/message", handler)

	log.Printf("🚀 MCP server listening on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
