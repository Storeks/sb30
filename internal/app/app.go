// Main block
package app

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/storeks/sb30/internal/handler"
)

// Entry point
func Run() {
	addr := ":4000"

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	rout := handler.NewHandler(infoLog, errorLog)
	rout.Init()
	srv := &http.Server{
		Addr:     addr,
		ErrorLog: errorLog,
		Handler:  rout.H,
	}

	infoLog.Println("Server start on localhost" + addr)
	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			errorLog.Fatal(err)
		}
	}()

	cancelChan := make(chan os.Signal, 1)
	signal.Notify(cancelChan, os.Interrupt)
	<-cancelChan

	infoLog.Println("Server shutdown ...")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	srv.Shutdown(ctx)

}
