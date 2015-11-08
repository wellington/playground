package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"text/template"
	"time"

	"github.com/gorilla/mux"
)

var certFile = "server.crt"
var keyFile = "server.key"

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

	svr.TLSConfig = &tls.Config{
		Certificates: make([]tls.Certificate, 1),
	}
	svr.TLSConfig.Certificates[0], err = tls.LoadX509KeyPair(
		certFile, keyFile)
	if err != nil {
		log.Fatal(err)
	}
	// Attaches http2
	// http2.ConfigureServer(svr, &http2.Server{})
	log.Println("Listening on:", lis.Addr())

	tlsLis := tls.NewListener(lis, svr.TLSConfig)
	log.Fatal(svr.Serve(tlsLis))
}

var tmpls *template.Template
var tplHome *template.Template
var cert tls.Certificate

func init() {
	var err error
	tmpls = template.New("nevermatch").Delims("{{{", "}}}")
	cert, err = tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		log.Fatal("cert failure:", err)
	}
}

func home(w http.ResponseWriter, r *http.Request) {
	// Put this here for easier development
	tmpls := template.New("nevermatch").Delims("{{{", "}}}")
	tplHome := template.Must(tmpls.ParseFiles("tmpl/index.html"))
	tplHome.ExecuteTemplate(w, "index.html", nil)
}

func doCompile(body io.Reader) (io.Reader, error) {
	cli := http.Client{}

	urlStr := "https://dumass.local:12345"
	slurp, err := cli.Post(urlStr, "application/json", body)
	if err != nil {
		return nil, err
	}
	fmt.Println("Received response as:", slurp.Proto)
	return slurp.Body, nil
}

func mustCompile(r io.Reader, err error) bytes.Buffer {
	if err != nil {
		panic(err)
	}
	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	if err != nil {
		panic(err)
	}
	return buf
}

func compile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	resp, err := doCompile(r.Body)
	if err != nil {
		log.Println("ERROR contacting wt:", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"could not contact upstream wellington"}`))
		return
	}
	io.Copy(w, resp)
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
