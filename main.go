package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	flag "github.com/spf13/pflag"
	"github.com/render-oss/render-mcp-server/cmd"
	"github.com/render-oss/render-mcp-server/pkg/cfg"
)

// 🧩 Struktura příchozí zprávy
type IncomingMessage struct {
	Sender  string `json:"sender"`
	Message string `json:"message"`
	Time    string `json:"time"`
}

// 🌐 URL Make webhooku (nahraď svojí aktuální URL z Make)
const makeWebhookURL = "https://hook.eu2.make.com/6fr8k32ac8ryvt6ickkxh55wkdjimwtf"

// 🧠 Handler pro příjem zprávy a odeslání do Make
func handleIncomingMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	var msg IncomingMessage
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	data, err := json.Marshal(msg)
	if err != nil {
		http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
		return
	}

	resp, err := http.Post(makeWebhookURL, "application/json", bytes.NewBuffer(data))
	if err != nil {
		http.Error(w, "Error sending to Make", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("📤 Odesláno do Make | Status: %s | Odpověď: %s\n", resp.Status, string(body))

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		http.Error(w, "Make returned error", http.StatusBadGateway)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("✅ Message forwarded to Make"))
}

// 🩵 Ping endpoint – ověření, že server běží
func handlePing(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("🏓 pong – Render MCP Server běží! ✅"))
}

func main() {
	versionFlag := flag.Bool("version", false, "Print version information and exit")
	flag.BoolVar(versionFlag, "v", false, "Print version information and exit")

	var transport string
	flag.StringVarP(&transport, "transport", "t", "", "Transport type (stdio or http)")
	flag.Parse()

	if transport == "" {
		if envTransport := os.Getenv("TRANSPORT"); envTransport != "" {
			transport = envTransport
		} else {
			transport = "stdio"
		}
	}

	if *versionFlag {
		fmt.Println("render-mcp-server version", cfg.Version)
		os.Exit(0)
	}

	// 🌍 Port z Renderu (hlavní server)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("🚀 Starting Render MCP Server with transport: %s\n", transport)
	fmt.Println("🔑 MAKE_WEBHOOK_TOKEN =", os.Getenv("MAKE_WEBHOOK_TOKEN"))
	fmt.Println("🔑 RENDER_API_KEY =", os.Getenv("RENDER_API_KEY"))
	fmt.Println("🔑 PORT =", port)
	fmt.Println("🔑 TRANSPORT =", os.Getenv("TRANSPORT"))

	// 🌐 Mini endpoint pro zprávy + ping – běží na 9090
	go func() {
		fmt.Println("🌐 Listening on http://localhost:9090 ...")
		http.HandleFunc("/message", handleIncomingMessage)
		http.HandleFunc("/ping", handlePing)
		if err := http.ListenAndServe(":9090", nil); err != nil {
			fmt.Println("❌ HTTP server error:", err)
		}
	}()

	// ▶️ Spusť MCP server (Render používá PORT z env)
	cmd.Serve(transport)
}
