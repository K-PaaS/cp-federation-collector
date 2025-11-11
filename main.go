package main

import (
	"context"
	"federation-metric-api/controller"
	_ "federation-metric-api/docs"
	"fmt"
	echoSwagger "github.com/swaggo/http-swagger"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var hostClusterName = "host-cluster"

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/actuator/health/liveness", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, "ok")
		})
		mux.HandleFunc("/actuator/health/readiness", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, "ready")
		})
		mux.Handle("/swagger/", echoSwagger.WrapHandler)

		http.ListenAndServe(":8001", mux)
	}()

	go controller.RepeatMetric(ctx)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	fmt.Println("Shutdown signal received.")
	//cancel()
	time.Sleep(2 * time.Second)
}
