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

type IncomingMessage struct {
	Sender  string `json:"sender"`
	Message string `json:"message"`
	Time    string `json:"time"`
}

const makeWebhookURL = "https://hook.eu2.make.com/6fr8k32ac8ryvt6ickkxh55wkdjimwtf"

func handleIncomingMessage(w http.ResponseWriter, r *http.Request) {
	fmt.Println("📩 /message hit")
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	var msg IncomingMessage
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		fmt.Println("❌ Invalid JSON:", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	fmt.Printf("➡️  From %s: %s (%s)\n", msg.Sender, msg.Message, msg.Time)

	data, err := json.Marshal(msg)
	if err != nil {
		fmt.Println("❌ JSON marshal:", err)
		http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
		return
	}

	resp, err := http.Post(makeWebhookURL, "application/json", bytes.NewBuffer(data))
	if err != nil {
		fmt.Println("❌ Send to Make:", err)
		http.Error(w, "Error sending to Make", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("📤 To Make | Status: %s | Body: %s\n", resp.Status, string(body))

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		http.Error(w, "Make returned error", http.StatusBadGateway)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("✅ Message forwarded to Make"))
}

func handlePing(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("🏓 pong – Render MCP Server běží a odpovídá ✅"))
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

	fmt.Println("🚀 Starting Render MCP Server with transport:", transport)
	fmt.Println("🔑 MAKE_WEBHOOK_TOKEN =", os.Getenv("MAKE_WEBHOOK_TOKEN"))
	fmt.Println("🔑 RENDER_API_KEY   =", os.Getenv("RENDER_API_KEY"))
	fmt.Println("🔑 TRANSPORT        =", os.Getenv("TRANSPORT"))

	// 👉 DŮLEŽITÉ: MCP přesuň na 9090, aby nezabral 8080
	os.Setenv("PORT", "9090")

	// Veřejný server na 8080 (přístupný z internetu)
	go func() {
		fmt.Println("🌐 Public server listening on port 8080 ...")
		http.HandleFunc("/ping", handlePing)
		http.HandleFunc("/message", handleIncomingMessage)
		if err := http.ListenAndServe(":8080", nil); err != nil {
			fmt.Println("❌ HTTP server error:", err)
		}
	}()

	// MCP server (interně) na 9090
	cmd.Serve(transport)
}
