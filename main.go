package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGTERM,
		os.Interrupt,
	)
	defer cancel()

	serv, err := NewServer("new.db", ServerOpts{
		ListenAddr:   ":4040",
		readTimeout:  time.Minute,
		writeTimeout: time.Minute,
	})
	if err != nil {
		panic(err)
	}
	go serv.Start(ctx)

	<-ctx.Done()
	slog.Info("shutting down server gracefully...")
}
