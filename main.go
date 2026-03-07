package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGTERM,
		os.Interrupt,
	)
	defer cancel()

	serv, err := NewServer(defaultDBFile, DefaultServerOpts())
	if err != nil {
		panic(err)
	}
	go serv.Start(ctx)

	<-ctx.Done()
	if err := serv.Shutdown(); err != nil {
		slog.Error("shutdown error", "err", err)
	}
}
