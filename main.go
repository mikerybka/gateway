package main

import (
	"crypto/tls"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/mikerybka/util"
	"golang.org/x/crypto/acme/autocert"
)

type Config map[string]struct {
	Port string `json:"port"`
	Auth struct {
		User string `json:"user"`
		Pass string `json:"pass"`
		Dir  string `json:"dir"`
	} `json:"auth"`
}

func main() {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "/root"
	}
	configPath := filepath.Join(home, "gateway.json")
	certPath := filepath.Join(home, "certs")
	b, err := os.ReadFile(configPath)
	if err != nil {
		panic(err)
	}
	conf := Config{}
	err = json.Unmarshal(b, &conf)
	if err != nil {
		panic(err)
	}
	hosts := []string{}
	mux := http.NewServeMux()
	for pattern, backend := range conf {
		p := parsePattern(pattern)
		hosts = append(hosts, p.Host)
		backendURL := "http://localhost:" + backend.Port
		u, err := url.Parse(backendURL)
		if err != nil {
			panic(err)
		}
		proxy := httputil.NewSingleHostReverseProxy(u)
		h := func(w http.ResponseWriter, r *http.Request) {
			// Auth
			if backend.Auth.Dir != "" {
				// TODO
			} else if backend.Auth.User != "" || backend.Auth.Pass != "" {
				user, pass, ok := r.BasicAuth()
				if !ok && backend.Auth.User != user || backend.Auth.Pass != pass {
					http.Error(w, "not authorized", http.StatusUnauthorized)
					return
				}
			}

			// Send req thru to backend
			r.URL.Path = strings.TrimPrefix(r.URL.Path, p.Path)
			if r.URL.Path == "" {
				r.URL.Path = "/"
			}
			proxy.ServeHTTP(w, r)
		}
		path := util.ParsePath(p.Path)
		if len(path) > 0 {
			mux.HandleFunc(pattern, h)
		}
		pattern += "/"
		mux.HandleFunc(pattern, h)
	}
	m := &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		Cache:      autocert.DirCache(certPath),
		HostPolicy: autocert.HostWhitelist(hosts...),
	}
	server := &http.Server{
		Addr: ":443",
		TLSConfig: &tls.Config{
			GetCertificate: m.GetCertificate,
		},
		Handler: mux,
	}
	go func() {
		http.ListenAndServe(":80", m.HTTPHandler(nil))
	}()
	log.Fatal(server.ListenAndServeTLS("", "")) // Cert and key are managed by autocert
}

type Pattern struct {
	Method string
	Host   string
	Path   string
}

func parsePattern(s string) *Pattern {
	parts := strings.Split(s, " ")
	p := &Pattern{}
	if len(parts) == 2 {
		p.Method = parts[0]
		p.Host, p.Path, _ = strings.Cut(parts[1], "/")
	} else {
		p.Host, p.Path, _ = strings.Cut(parts[0], "/")
	}
	p.Path = "/" + p.Path
	return p
}
