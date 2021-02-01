package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"https-proxy-server/mydb"
	"io"
	"log"
	"net"
	"net/http"
	"time"
)

//--------------------------
//--------------------------

func addRequestToDb(req *http.Request) {
	url := fmt.Sprintf("%v %v %v", req.Method, req.URL, req.Proto)
	data := mydb.Record{ReqURL: url, Port: req.Proto}
	mydb.InsertRecord(data)
}

//--------------------------
//			Handle HTTPS
//--------------------------
func handleTuneling(w http.ResponseWriter, r *http.Request) {
	// fmt.Println(w)
	addRequestToDb(r)
	destConn, err := net.DialTimeout("tcp", r.Host, 10*time.Second)
	// fmt.Println(destConn.LocalAddr().String())
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}
	clientConn, _, err := hijacker.Hijack()
	// fmt.Println(clientConn.LocalAddr().String())
	// fmt.Println(clientConn.RemoteAddr().String())
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
	}

	go transferTo(destConn, clientConn)
	go transferFrom(clientConn, destConn)

}

//--------------------------
//			Handle HTTP
//--------------------------
func handleHTTP(w http.ResponseWriter, r *http.Request) {
	resp, err := http.DefaultTransport.RoundTrip(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()
	copyHeader(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

//--------------------------
// 			Transfer Data
//--------------------------
func transferTo(destination io.WriteCloser, source io.ReadCloser) {
	defer destination.Close()
	defer source.Close()
	// io.Copy(destination, source)
	// if wt, ok := source.Read([]byte("asdasdasd"), ok {
	// 	destination.Write(wt)
	// }

	size, _ := io.Copy(destination, source)
	fmt.Println(size)
}

func transferFrom(destination io.WriteCloser, source io.ReadCloser) {
	defer destination.Close()
	defer source.Close()
	io.Copy(destination, source)
}

//--------------------------
//			Copy Header
//--------------------------
func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

//*************************
//				 Main
//*************************
func main() {
	var pemPath string
	var keyPath string
	var proto string
	flag.StringVar(&pemPath, "pem", "server.pem", "Path to pem file.")
	flag.StringVar(&keyPath, "key", "server.key", "Path to key file.")
	flag.StringVar(&proto, "proto", "https", "Proxy protocol (http or https)")

	mydb.Init()

	server := &http.Server{
		Addr: ":8888",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodConnect {
				handleTuneling(w, r)
			} else {
				handleHTTP(w, r)
			}
		}),
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}

	if proto == "http" {
		log.Fatal(server.ListenAndServe())
	} else {
		log.Fatal(server.ListenAndServeTLS(pemPath, keyPath))
	}
}
