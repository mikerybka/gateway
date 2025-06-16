package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/acme/autocert"
)

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
	conf := map[string]string{}
	err = json.Unmarshal(b, &conf)
	if err != nil {
		panic(err)
	}
	hosts := []string{}
	mux := http.NewServeMux()
	for pattern, backendURL := range conf {
		host := strings.Split(pattern, "/")[0]
		hosts = append(hosts, host)
		u, err := url.Parse(backendURL)
		if err != nil {
			panic(err)
		}
		mux.Handle(fmt.Sprintf("%s/", pattern), httputil.NewSingleHostReverseProxy(u))
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
