package main

import (
	"crypto/tls"
	"log"
	"net"
	"net/http"

	"golang.org/x/net/http2"

	"github.com/gorilla/mux"
)

func main() {

	r := mux.NewRouter()
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

func home(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Home is where the handler is"))
}
