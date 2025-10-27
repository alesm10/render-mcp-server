package main

import (
	"bytes"
	"encoding/json"
	"fmt"
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

// 🌐 URL Make webhooku (nahraď svojí URL z Make)
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

	fmt.Printf("📤 Message from %s sent to Make (status %s)\n", msg.Sender, resp.Status)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("✅ Message forwarded to Make"))
}

func main() {
	// 🏁 Definice a načtení flagů
	versionFlag := flag.Bool("version", false, "Print version information and exit")
	flag.BoolVar(versionFlag, "v", false, "Print version information and exit")

	var transport string
	flag.StringVarP(&transport, "transport", "t", "", "Transport type (stdio or http)")
	flag.Parse()

	// 🔧 Transport z ENV
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

	// 🚀 Start info
	fmt.Printf("🚀 Starting Render MCP Server with transport: %s\n", transport)
	fmt.Println("🔑 MAKE_WEBHOOK_TOKEN =", os.Getenv("MAKE_WEBHOOK_TOKEN"))
	fmt.Println("🔑 RENDER_API_KEY =", os.Getenv("RENDER_API_KEY"))
	fmt.Println("🔑 PORT =", os.Getenv("PORT"))
	fmt.Println("🔑 TRANSPORT =", os.Getenv("TRANSPORT"))

	// 🌍 Spusť mini HTTP endpoint paralelně
	go func() {
		fmt.Println("🌐 Listening on http://localhost:8090/message")
		http.HandleFunc("/message", handleIncomingMessage)
		if err := http.ListenAndServe(":8090", nil); err != nil {
			fmt.Println("❌ HTTP server error:", err)
		}
	}()

	// ▶️ Spusť Render MCP server
	cmd.Serve(transport)
}
