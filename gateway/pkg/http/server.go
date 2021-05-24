package http

import (
	"context"
	"log"
	"net/http"
	"time"
)

// ListenAndServe serves an http server over TCP handling graceful shutdown
func ListenAndServe(ctx context.Context, srv *http.Server) error {
	go func() {
		<-ctx.Done()
		shutdownContext, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		err := srv.Shutdown(shutdownContext)
		if err != nil {
			log.Fatal(err)
		}
	}()
	log.Printf("start listening for HTTP requests on " + srv.Addr)
	return srv.ListenAndServe()
}
