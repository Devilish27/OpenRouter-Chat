package main

import (
	_ "embed"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

//go:embed index.html
var page []byte

var (
	dataDir, chatsDir, port string
	lastPing                atomic.Int64
	idRe                    = regexp.MustCompile(`^[A-Za-z0-9_-]{1,64}$`)
)

func main() {
	exe, _ := os.Executable()
	dataDir = filepath.Join(filepath.Dir(exe), "data")
	chatsDir = filepath.Join(dataDir, "chats")
	os.MkdirAll(chatsDir, 0o755)

	// Find a free port, or hand over to an already-running instance.
	var ln net.Listener
	for p := 47821; p < 47831; p++ {
		addr := "127.0.0.1:" + strconv.Itoa(p)
		l, err := net.Listen("tcp", addr)
		if err == nil {
			ln, port = l, strconv.Itoa(p)
			break
		}
		if isOurs(addr) {
			openBrowser("http://" + addr)
			return
		}
	}
	if ln == nil {
		return
	}

	// Exit by itself when the browser page is gone (no heartbeat for 150 s).
	lastPing.Store(time.Now().Unix())
	go func() {
		for range time.Tick(5 * time.Second) {
			if time.Now().Unix()-lastPing.Load() > 150 {
				os.Exit(0)
			}
		}
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(page)
	})
	mux.HandleFunc("/api/ping", func(w http.ResponseWriter, r *http.Request) {
		lastPing.Store(time.Now().Unix())
		w.Write([]byte("openrouter-chat-ok"))
	})
	mux.HandleFunc("/api/quit", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "method not allowed", 405)
			return
		}
		w.WriteHeader(204)
		go func() { time.Sleep(300 * time.Millisecond); os.Exit(0) }()
	})
	mux.HandleFunc("/api/settings", func(w http.ResponseWriter, r *http.Request) {
		f := filepath.Join(dataDir, "settings.json")
		if r.Method == "PUT" {
			put(w, r, f, 1<<20)
			return
		}
		b, err := os.ReadFile(f)
		if err != nil {
			b = []byte("{}")
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(b)
	})
	mux.HandleFunc("/api/chats", list)
	mux.HandleFunc("/api/chats/", one)

	go openBrowser("http://127.0.0.1:" + port)
	http.Serve(ln, guard(mux))
}

// guard rejects requests from other sites (DNS rebinding / cross-origin).
func guard(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		okHost := r.Host == "127.0.0.1:"+port || r.Host == "localhost:"+port
		o := r.Header.Get("Origin")
		if !okHost || (o != "" && o != "http://"+r.Host) {
			http.Error(w, "forbidden", 403)
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		h.ServeHTTP(w, r)
	})
}

type meta struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Updated string `json:"updated"`
}

func list(w http.ResponseWriter, r *http.Request) {
	es, _ := os.ReadDir(chatsDir)
	out := []meta{}
	for _, e := range es {
		if !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		b, err := os.ReadFile(filepath.Join(chatsDir, e.Name()))
		if err != nil {
			continue
		}
		var m meta
		if json.Unmarshal(b, &m) == nil && m.ID != "" {
			out = append(out, m)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Updated > out[j].Updated })
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)
}

func one(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/chats/")
	if !idRe.MatchString(id) {
		http.Error(w, "bad id", 400)
		return
	}
	f := filepath.Join(chatsDir, id+".json")
	switch r.Method {
	case "GET":
		b, err := os.ReadFile(f)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(b)
	case "PUT":
		put(w, r, f, 50<<20)
	case "DELETE":
		os.Remove(f)
		w.WriteHeader(204)
	default:
		http.Error(w, "method not allowed", 405)
	}
}

// put validates JSON and writes it atomically (temp file, then rename).
func put(w http.ResponseWriter, r *http.Request, f string, max int64) {
	b, err := io.ReadAll(io.LimitReader(r.Body, max))
	if err != nil || !json.Valid(b) {
		http.Error(w, "bad json", 400)
		return
	}
	tmp := f + ".tmp"
	if os.WriteFile(tmp, b, 0o600) != nil || os.Rename(tmp, f) != nil {
		http.Error(w, "write failed", 500)
		return
	}
	w.WriteHeader(204)
}

func isOurs(addr string) bool {
	c := http.Client{Timeout: time.Second}
	r, err := c.Get("http://" + addr + "/api/ping")
	if err != nil {
		return false
	}
	defer r.Body.Close()
	b, _ := io.ReadAll(io.LimitReader(r.Body, 64))
	return string(b) == "openrouter-chat-ok"
}

func openBrowser(u string) {
	if os.Getenv("NOOPEN") != "" {
		return
	}
	switch runtime.GOOS {
	case "windows":
		exec.Command("rundll32", "url.dll,FileProtocolHandler", u).Start()
	case "darwin":
		exec.Command("open", u).Start()
	default:
		exec.Command("xdg-open", u).Start()
	}
}
