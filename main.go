package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	flag "github.com/spf13/pflag"
)

// Struktura příchozí zprávy
type IncomingMessage struct {
	Sender  string `json:"sender"`
	Message string `json:"message"`
	Time    string `json:"time"`
}

// 🔗 Webhook URL z Make.com (nezapomeň, že to je tvůj vlastní)
const makeWebhookURL = "https://hook.eu2.make.com/t85w0984wnlyu7oklq8kdnttsj90iz6n"

// Handler pro příchozí zprávy
func handleIncomingMessage(w http.ResponseWriter, r *http.Request) {
	fmt.Println("📩 /message hit")

	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	var msg IncomingMessage
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		fmt.Println("❌ Invalid JSON:", err)
		return
	}

	fmt.Printf("➡️ Received message: %+v\n", msg)

	// Připrav JSON payload pro Make
	payload, err := json.Marshal(msg)
	if err != nil {
		http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
		fmt.Println("❌ Error encoding JSON:", err)
		return
	}

	fmt.Println("🌍 Sending to Make:", makeWebhookURL)
	resp, err := http.Post(makeWebhookURL, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		http.Error(w, "Error sending to Make", http.StatusBadGateway)
		fmt.Println("❌ Make returned error:", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Println("⬅️ Make response status:", resp.Status)
	fmt.Println("⬅️ Make response body:", string(body))

	if resp.StatusCode != 200 {
		fmt.Println("⚠️ Make returned non-200, status:", resp.Status)
		http.Error(w, "Make returned non-200", http.StatusBadGateway)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("✅ Message forwarded to Make"))
	fmt.Println("✅ Successfully forwarded to Make!")
}

// Hlavní funkce
func main() {
	var transport string
	flag.StringVarP(&transport, "transport", "t", "", "Transport type (stdio or http)")
	flag.Parse()

	// 🧩 Render přiřazuje port dynamicky
	port := os.Getenv("PORT")
	if port == "" {
		port = "10000" // fallback
	}
	fmt.Println("📦 PORT env var:", port)

	http.HandleFunc("/message", handleIncomingMessage)

	fmt.Println("🚀 MCP server listening on port", port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		fmt.Println("❌ Server error:", err)
	}
}
