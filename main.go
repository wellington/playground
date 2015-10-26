package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"text/template"
	"time"

	"golang.org/x/net/http2"

	"github.com/gorilla/mux"
)

func main() {

	r := mux.NewRouter()
	r.HandleFunc("/stream", stream)
	r.HandleFunc("/compile", compile)

	// serve static files
	ui := http.FileServer(http.Dir("./static/"))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", ui))

	// r.Handle("/public/", http.StripPrefix("/public/", ui))
	r.HandleFunc("/", home)

	http.Handle("/", r)

	lis, err := net.Listen("tcp", "127.0.0.1:6464")
	if err != nil {
		log.Fatal(err)
	}
	svr := &http.Server{}
	certFile := "server.crt"
	keyFile := "server.key"

	svr.TLSConfig = &tls.Config{
		Certificates: make([]tls.Certificate, 1),
	}
	svr.TLSConfig.Certificates[0], err = tls.LoadX509KeyPair(
		certFile, keyFile)
	if err != nil {
		log.Fatal(err)
	}
	// Attaches http2
	http2.ConfigureServer(svr, &http2.Server{})
	log.Println("Listening on:", lis.Addr())

	tlsLis := tls.NewListener(lis, svr.TLSConfig)
	log.Fatal(svr.Serve(tlsLis))
}

var tmpls *template.Template
var tplHome *template.Template

func init() {
	tmpls = template.New("nevermatch").Delims("{{{", "}}}")
}

func home(w http.ResponseWriter, r *http.Request) {
	// Put this here for easier development
	tmpls := template.New("nevermatch").Delims("{{{", "}}}")
	tplHome := template.Must(tmpls.ParseFiles("tmpl/index.html"))
	tplHome.ExecuteTemplate(w, "index.html", nil)
}

func compile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	slurp, err := http.Post("http://localhost:12345", "", r.Body)
	if err != nil {
		log.Println("ERROR contacting wt:", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"could not contact upstream wellington"}`))
		return
	}
	io.Copy(w, slurp.Body)
}

func stream(w http.ResponseWriter, r *http.Request) {
	clientGone := w.(http.CloseNotifier).CloseNotify()
	w.Header().Set("Content-Type", "text/html")
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	fmt.Fprintf(w, "<!-- # ~1KB of junk to force browsers to start rendering immediately: \n -->")
	io.WriteString(w, strings.Repeat("<!--# xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx\n-->", 13))

	for {
		fmt.Fprintf(w, "<p>%v</p>", time.Now())
		w.(http.Flusher).Flush()
		select {
		case <-ticker.C:
		case <-clientGone:
			log.Printf("Client %v disconnected from the clock", r.RemoteAddr)
			return
		}
	}
}
