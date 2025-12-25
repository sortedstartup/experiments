package main

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
)

func main() {

	// custom log level
	TRACE := slog.Level(slog.LevelDebug + 1)

	//-- slog init starts
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true, // behaviour in go build binary - print full source when built
		Level:     TRACE,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.SourceKey {
				a.Value = slog.StringValue(filepath.Base(a.Value.String()))
			}
			return a
		},
	})

	logger := slog.New(handler)
	slog.SetDefault(logger)
	//-- Slog init ends
	//
	ctx := context.Background()

	slog.Log(ctx, slog.LevelInfo, "info level message")
	slog.Log(ctx, TRACE, "debug+1 level message")

}
