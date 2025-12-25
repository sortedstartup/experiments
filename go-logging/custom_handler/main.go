package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/charmbracelet/lipgloss"
)

func main() {

	// custom log level
	TRACE := slog.Level(slog.LevelDebug + 1)

	//-- slog init starts
	// handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
	// 	AddSource: true, // behaviour in go build binary - print full source when built
	// 	Level:     TRACE,
	// })
	handler := NewCustomHandler()

	logger := slog.New(handler)
	slog.SetDefault(logger)
	//-- Slog init ends
	//
	ctx := context.Background()

	slog.Log(ctx, slog.LevelInfo, "info level message")
	slog.Log(ctx, slog.LevelDebug, "debug level message")
	slog.Log(ctx, slog.LevelError, "message level message")
	slog.Log(ctx, TRACE, "debug+1 level message")

}

type CustomHandler struct {
	timeStyle lipgloss.Style

	levelStrings map[slog.Level]string

	levelNameStyles map[slog.Level]lipgloss.Style
	levelStyles     map[slog.Level]lipgloss.Style
}

func NewCustomHandler() CustomHandler {

	c := CustomHandler{}

	c.levelStrings = map[slog.Level]string{
		slog.LevelInfo:                  "INFO",
		slog.LevelDebug:                 "DEBUG",
		slog.LevelError:                 "ERROR",
		slog.Level(slog.LevelDebug + 1): "TRACE",
	}

	c.timeStyle = lipgloss.NewStyle()

	c.levelNameStyles = make(map[slog.Level]lipgloss.Style)
	c.levelStyles = make(map[slog.Level]lipgloss.Style)

	c.levelNameStyles[slog.LevelInfo] = lipgloss.NewStyle().
		Bold(true).
		Width(5).
		Align(lipgloss.Center).
		// AlignHorizontal(lipgloss.Center).
		Foreground(lipgloss.Color("#222222")) // dark text
		// Background(lipgloss.Color("#A7FFEB"))  // light teal

	c.levelNameStyles[slog.LevelDebug] = lipgloss.NewStyle().
		Bold(true).
		Width(5).
		Foreground(lipgloss.Color("#37474F")). // dark blue-grey
		Background(lipgloss.Color("#B0BEC5"))  // light blue-grey

	c.levelNameStyles[slog.Level(slog.LevelDebug+1)] = lipgloss.NewStyle().
		Bold(true).
		Width(5).
		Foreground(lipgloss.Color("#263238")). // blue-grey text
		Background(lipgloss.Color("#80CBC4"))  // teal

	c.levelNameStyles[slog.LevelError] = lipgloss.NewStyle().
		Bold(true).
		Width(5).
		Foreground(lipgloss.Color("#FAFAFA")). // white text
		Background(lipgloss.Color("#FF1744"))  // bright red

	c.levelStyles[slog.LevelInfo] = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA")) // white text

	c.levelStyles[slog.LevelDebug] = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA")) // white text

	c.levelStyles[slog.LevelError] = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FF1744")) // white text

	c.levelStyles[slog.Level(slog.LevelDebug+1)] = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")) // white text

	return c
}

// Enabled reports whether the handler handles records at the given level.
// The handler ignores records whose level is lower.
// It is called early, before any arguments are processed,
// to save effort if the log event should be discarded.
// If called from a Logger method, the first argument is the context
// passed to that method, or context.Background() if nil was passed
// or the method does not take a context.
// The context is passed so Enabled can use its values
// to make a decision.
func (xc CustomHandler) Enabled(c context.Context, l slog.Level) bool {
	//	fmt.Printf("Enabled %d\n", l)
	return true

}

// Handle handles the Record.
// It will only be called when Enabled returns true.
// The Context argument is as for Enabled.
// It is present solely to provide Handlers access to the context's values.
// Canceling the context should not affect record processing.
// (Among other things, log messages may be necessary to debug a
// cancellation-related problem.)
//
// Handle methods that produce output should observe the following rules:
//   - If r.Time is the zero time, ignore the time.
//   - If r.PC is zero, ignore it.
//   - Attr's values should be resolved.
//   - If an Attr's key and value are both the zero value, ignore the Attr.
//     This can be tested with attr.Equal(Attr{}).
//   - If a group's key is empty, inline the group's Attrs.
//   - If a group has no Attrs (even if it has a non-empty key),
//     ignore it.
//
// [Logger] discards any errors from Handle. Wrap the Handle method to
// process any errors from Handlers.
func (c CustomHandler) Handle(ctx context.Context, r slog.Record) error {

	levelNameStyle, ok := c.levelNameStyles[r.Level]
	if !ok {
		levelNameStyle = c.levelNameStyles[slog.LevelInfo]
	}

	levelTextStyle, ok := c.levelStyles[r.Level]
	if !ok {
		levelNameStyle = c.levelStyles[slog.LevelInfo]
	}

	// Map from slog.Level to string at struct level
	levelStr, ok := c.levelStrings[r.Level]
	if !ok {
		levelStr = "UNKNOWN"
	}

	fmt.Println(c.timeStyle.Render(r.Time.Format(time.Stamp)) + " " + levelNameStyle.Render(levelStr) + " " + levelTextStyle.Render(r.Message))
	return nil
}

// WithAttrs returns a new Handler whose attributes consist of
// both the receiver's attributes and the arguments.
// The Handler owns the slice: it may retain, modify or discard it.
func (c CustomHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	fmt.Printf("WithAttrs %+v\n", attrs)
	return c
}

// WithGroup returns a new Handler with the given group appended to
// the receiver's existing groups.
// The keys of all subsequent attributes, whether added by With or in a
// Record, should be qualified by the sequence of group names.
//
// How this qualification happens is up to the Handler, so long as
// this Handler's attribute keys differ from those of another Handler
// with a different sequence of group names.
//
// A Handler should treat WithGroup as starting a Group of Attrs that ends
// at the end of the log event. That is,
//
//	logger.WithGroup("s").LogAttrs(ctx, level, msg, slog.Int("a", 1), slog.Int("b", 2))
//
// should behave like
//
//	logger.LogAttrs(ctx, level, msg, slog.Group("s", slog.Int("a", 1), slog.Int("b", 2)))
//
// If the name is empty, WithGroup returns the receiver.
func (c CustomHandler) WithGroup(name string) slog.Handler {
	fmt.Printf("WithGroup %s\n", name)
	return c
}
