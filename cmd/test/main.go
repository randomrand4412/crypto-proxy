package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

func getSign(rw http.ResponseWriter, req *http.Request) {
	fmt.Println("Hello all from sign")
	fmt.Println(req.Header["Authorization"])
	fmt.Println(req.URL.Query())

	if !req.URL.Query().Has("message") {
		http.Error(rw, "invalid request", http.StatusBadRequest)
		return
	}

	sleepTime := time.Duration(rand.Intn(10000)) * time.Millisecond
	fmt.Printf("sleeping for %d\n", sleepTime)

	time.Sleep(sleepTime)

	rw.WriteHeader(http.StatusOK)
	fmt.Fprint(rw, "2312312")
}

func getVerify(rw http.ResponseWriter, req *http.Request) {
	fmt.Println("Hello all from verify")
	fmt.Println(req.Header["Authorization"])
	fmt.Println(req.URL.Query())

	if !(req.URL.Query().Has("message") && req.URL.Query().Has("signature")) {
		http.Error(rw, "invalid request", http.StatusBadRequest)
		return
	}

	rw.WriteHeader(http.StatusOK)
	fmt.Fprint(rw, "")
}

func main() {
	http.Handle("/crypto/sign", http.HandlerFunc(getSign))
	http.Handle("/crypto/verify", http.HandlerFunc(getVerify))

	http.ListenAndServe(":9090", nil)
}
