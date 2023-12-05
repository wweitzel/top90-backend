package jsonlogger

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"sync"
)

const (
	timeFormat = "[15:04:05.000]"
	reset      = "\033[0m"

	black        = 30
	red          = 31
	green        = 32
	yellow       = 33
	blue         = 34
	magenta      = 35
	cyan         = 36
	lightGray    = 37
	darkGray     = 90
	lightRed     = 91
	lightGreen   = 92
	lightYellow  = 93
	lightBlue    = 94
	lightMagenta = 95
	lightCyan    = 96
	white        = 97
)

func New(opts *Options) *slog.Logger {
	return slog.New(newHandler(opts))
}

type handler struct {
	h    slog.Handler
	b    *bytes.Buffer
	m    *sync.Mutex
	opts Options
}

type Options struct {
	Level    slog.Leveler
	Colorize bool
}

func newHandler(opts *Options) *handler {
	if opts == nil {
		opts = &Options{}
	}
	if opts.Level == nil {
		opts.Level = slog.LevelInfo
	}
	slogOpts := &slog.HandlerOptions{Level: opts.Level}

	b := &bytes.Buffer{}
	return &handler{
		b: b,
		h: slog.NewJSONHandler(b, &slog.HandlerOptions{
			Level:       slogOpts.Level,
			AddSource:   slogOpts.AddSource,
			ReplaceAttr: replaceAttr(slogOpts.ReplaceAttr),
		}),
		m:    &sync.Mutex{},
		opts: *opts,
	}
}

func (h *handler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.h.Enabled(ctx, level)
}

func (h *handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &handler{h: h.h.WithAttrs(attrs), b: h.b, m: h.m}
}

func (h *handler) WithGroup(name string) slog.Handler {
	return &handler{h: h.h.WithGroup(name), b: h.b, m: h.m}
}

func (h *handler) Handle(ctx context.Context, r slog.Record) error {
	level := r.Level.String() + ":"

	switch r.Level {
	case slog.LevelDebug:
		level = h.colorize(lightGray, level)
	case slog.LevelInfo:
		level = h.colorize(cyan, level)
	case slog.LevelWarn:
		level = h.colorize(yellow, level)
	case slog.LevelError:
		level = h.colorize(red, level)
	default:
		level = h.colorize(lightGray, level)
	}

	attrs, err := h.unmarshalAttrs(ctx, r)
	if err != nil {
		return err
	}
	bytes, err := json.MarshalIndent(attrs, "", "  ")
	if err != nil {
		return fmt.Errorf("error when marshaling attrs: %w", err)
	}
	if len(attrs) == 0 {
		bytes = nil
	}

	fmt.Println(
		h.colorize(lightGray, r.Time.Format(timeFormat)),
		level,
		h.colorize(white, r.Message),
		h.colorize(darkGray, string(bytes)),
	)
	return nil
}

func (h *handler) unmarshalAttrs(
	ctx context.Context,
	r slog.Record,
) (map[string]any, error) {
	h.m.Lock()
	defer func() {
		h.b.Reset()
		h.m.Unlock()
	}()
	if err := h.h.Handle(ctx, r); err != nil {
		return nil, fmt.Errorf("error unmarshalling attributes: %w", err)
	}

	var attrs map[string]any
	err := json.Unmarshal(h.b.Bytes(), &attrs)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling attributes: %w", err)
	}
	return attrs, nil
}

func (h *handler) colorize(colorCode int, v string) string {
	if h.opts.Colorize {
		return fmt.Sprintf("\033[%sm%s%s", strconv.Itoa(colorCode), v, reset)
	}
	return v
}

func replaceAttr(
	next func([]string, slog.Attr) slog.Attr,
) func([]string, slog.Attr) slog.Attr {
	return func(groups []string, a slog.Attr) slog.Attr {
		if a.Key == slog.TimeKey ||
			a.Key == slog.LevelKey ||
			a.Key == slog.MessageKey {
			return slog.Attr{}
		}
		if next == nil {
			return a
		}
		return next(groups, a)
	}
}
