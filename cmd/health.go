package cmd

import (
    "encoding/json"
    "net/http"
)

func StartHealthEndpoint() {
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]interface{}{
            "jsonrpc": "2.0",
            "id":      1,
            "result": map[string]interface{}{
                "ok":      true,
                "message": "render-mcp-server running",
            },
        })
    })

    go http.ListenAndServe(":8080", nil)
}
