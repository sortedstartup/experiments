package main

import (
	"context"
	"log/slog"
	"os"
)

func main() {

	// custom log level
	TRACE := slog.Level(slog.LevelDebug + 1)

	//-- slog init starts
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     TRACE,
	})

	logger := slog.New(handler)
	slog.SetDefault(logger)
	//-- Slog init ends
	//
	ctx := context.Background()

	slog.Log(ctx, slog.LevelInfo, "info level message")
	slog.Log(ctx, TRACE, "debug+1 level message")

}
