package main

import (
	"flag"
	"fmt"
	"net/http"

	control "mygolibs/control"
)

// A simplest example of parsing command args
var (
	Port int
)

// Parse finished, now you can use `Port` directly.
func init() {
	flag.IntVar(&Port, "port", 8045, "Running port")
	flag.Parse()
}

func hello(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "hello\n")
}
func placeholder(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "test is not finished yet...\n")
}

func controlExec(w http.ResponseWriter, req *http.Request) {
	res, err := control.Exec("ls -a")
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}
	fmt.Fprint(w, res)
}

func headers(w http.ResponseWriter, req *http.Request) {
	for name, headers := range req.Header {
		for _, h := range headers {
			fmt.Fprintf(w, "%v: %v\n", name, h)
		}
	}
}

func main() {
	// examples
	http.HandleFunc("/hello", hello)
	http.HandleFunc("/headers", headers)

	// module tester
	http.HandleFunc("/test/control/exec", controlExec)
	http.HandleFunc("/test/control/metric/static", controlExec)
	http.HandleFunc("/test/control/metric/dynamic", controlExec)

	http.HandleFunc("/test/encrypt/hmac", placeholder)

	http.HandleFunc("/test/files/yaml", placeholder)

	// TODO: how to test grpc???
	// http.HandleFunc("/test/grpc/node", placeholder)

	// serve
	http.ListenAndServe(fmt.Sprintf(":%d", Port), nil)
}
