package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func handler(ctx context.Context) http.Handler {
	setTimeChan := make(chan *time.Time)
	getTimeChan := make(chan chan *time.Time)

	now := time.Now()

	go func() {
		savedTime := &now

		for {
			select {
			case <-ctx.Done():
				break
			case t := <- setTimeChan:
				savedTime = t
			case ch := <- getTimeChan:
				ch <- savedTime
			}
		}
	}()

	router := http.NewServeMux()

	router.HandleFunc("GET /time", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")

		ch := make(chan *time.Time)
		getTimeChan <- ch

		time := <- ch

		w.Write([]byte(strconv.Itoa(int(time.UnixMilli()))))
	})

	router.HandleFunc("POST /time", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "text/plain" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		i, err := strconv.Atoi(string(body))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		time := time.UnixMilli(int64(i))
		setTimeChan <- &time
	})

	return router
}

func server(ctx context.Context, ready chan bool) {
	router := handler(ctx)

	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	ready <- true

	server := http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	go func() {
		err = server.Serve(listener)
		if err != nil {
			log.Fatal(err)
		}
	}()

	<-ctx.Done()

	err = server.Shutdown(ctx)
	if err != nil {
		log.Fatal(err)
	}
}

func client() {
	_, err := http.Post("http://127.0.0.1:8080/time", "text/plain", strings.NewReader("123456789"))
	if err != nil {
		log.Fatal(err)
	}

	resp, err := http.Get("http://127.0.0.1:8080/time")
	if err != nil {
		log.Fatal(err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	fmt.Println(string(body))
}

func main() {
	ready := make(chan bool, 1)

	go server(context.Background(), ready)
	<-ready

	client()
}
