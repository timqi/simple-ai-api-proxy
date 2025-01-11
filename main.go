package main

import (
	"flag"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"os/exec"
	"strconv"
)

const (
	DAEMON         = "daemon"
	OPENAI_HOST    = "api.openai.com"
	ANTHROPIC_HOST = "api.anthropic.com"
)

var (
	port         int
	damaen       bool
	openaiKey    string
	anthropicKey string
	code         string
)

func init() {
	flag.IntVar(&port, "port", 8080, "Listen port")
	flag.BoolVar(&damaen, DAEMON, false, "Run in background")
	flag.StringVar(&openaiKey, "openai-key", "", "OpenAI API Key")
	flag.StringVar(&anthropicKey, "anthropic-key", "", "Anthropic API Key")
	flag.StringVar(&code, "code", "", "Access code for proxy")
}

func ReverseProxyHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[*] receive a request from %s, request header: %s: \n", r.RemoteAddr, r.Header)

	if code != "" {
		authHeader := r.Header.Get("Authorization")
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			authHeader = authHeader[7:]
		}
		apiKeyHeader := r.Header.Get("X-Api-Key")

		if authHeader != code && apiKeyHeader != code {
			http.Error(w, "Invalid access code", http.StatusUnauthorized)
			return
		}
	}

	var target string
	if r.URL.Path[:7] == "/openai" {
		target = OPENAI_HOST
		r.URL.Path = r.URL.Path[7:]
	} else if r.URL.Path[:10] == "/anthropic" {
		target = ANTHROPIC_HOST
		r.URL.Path = r.URL.Path[10:]
	} else {
		http.Error(w, "Invalid endpoint", http.StatusBadRequest)
		return
	}

	director := func(req *http.Request) {
		req.URL.Scheme = "https"
		req.URL.Host = target
		req.Host = target

		if target == ANTHROPIC_HOST {
			req.Header.Del("Authorization")
			if anthropicKey != "" {
				req.Header.Set("X-API-Key", anthropicKey)
			} else {
				auth := r.Header.Get("Authorization")
				if len(auth) > 7 && auth[:7] == "Bearer " {
					req.Header.Set("X-API-Key", auth[7:])
				} else {
					req.Header.Set("X-API-Key", auth)
				}
			}
		} else {
			if openaiKey != "" {
				req.Header.Set("Authorization", "Bearer "+openaiKey)
			} else {
				req.Header.Set("Authorization", r.Header.Get("Authorization"))
			}
		}
	}

	proxy := &httputil.ReverseProxy{Director: director}
	proxy.ServeHTTP(w, r)
	log.Printf("[*] receive the destination website response header: %s\n", w.Header())
}

func StripSlice(slice []string, element string) []string {
	for i := 0; i < len(slice); {
		if slice[i] == element && i != len(slice)-1 {
			slice = append(slice[:i], slice[i+1:]...)
		} else if slice[i] == element && i == len(slice)-1 {
			slice = slice[:i]
		} else {
			i++
		}
	}
	return slice
}

func SubProcess(args []string) *exec.Cmd {
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	if err != nil {
		log.Printf("[-] Error: %s\n", err)
	}
	return cmd
}

func main() {
	flag.Parse()
	log.Printf("[*] PID: %d PPID: %d ARG: %s\n", os.Getpid(), os.Getppid(), os.Args)
	if damaen {
		SubProcess(StripSlice(os.Args, "-"+DAEMON))
		log.Printf("[*] Daemon running in PID: %d PPID: %d\n", os.Getpid(), os.Getppid())
		os.Exit(0)
	}
	log.Printf("[*] Forever running in PID: %d PPID: %d\n", os.Getpid(), os.Getppid())
	log.Printf("[*] Starting server at port %v\n", port)
	if err := http.ListenAndServe(":"+strconv.Itoa(port), http.HandlerFunc(ReverseProxyHandler)); err != nil {
		log.Fatal(err)
	}
}
