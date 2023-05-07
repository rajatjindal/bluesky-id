package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	spinhttp "github.com/fermyon/spin/sdk/go/http"
)

const (
	base          = "https://bsky.social/xrpc"
	resolveHandle = base + "/com.atproto.identity.resolveHandle"
)

func init() {
	spinhttp.Handle(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			ui(w, r)
			return
		}

		if r.URL.Path == "/api" {
			backend(w, r)
			return
		}

		http.Error(w, "not found", http.StatusNotFound)
	})
}

//go:embed index.html
var indexPage string

func ui(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("content-type", "text/html")
	fmt.Fprint(w, indexPage)
}

func backend(w http.ResponseWriter, r *http.Request) {
	handle := r.URL.Query().Get("handle")
	handle, _ = strings.CutPrefix(handle, "@")
	if handle == "" {
		http.Error(w, "handle is required as query param", http.StatusBadRequest)
		return
	}

	resp, err := http.Get(fmt.Sprintf("%s?handle=%s", resolveHandle, handle))
	if err != nil {
		fmt.Println("ERROR ", err.Error())
		http.Error(w, "failed to make request", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("ERROR ", err.Error())
		http.Error(w, "failed to read response", http.StatusInternalServerError)
		return
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("ERROR expected code %d, got %d. body: %s", http.StatusOK, resp.StatusCode, string(raw))
		http.Error(w, "failed to make request", http.StatusInternalServerError)
		return
	}

	didResp := struct {
		DID string `json:"did"`
	}{}

	err = json.Unmarshal(raw, &didResp)
	if err != nil {
		fmt.Println("ERROR ", err.Error())
		http.Error(w, "failed to parse response", http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, didResp.DID)
}

func main() {}
