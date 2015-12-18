package gojilogger

import (
	"bytes"
	"log"
	"net/http"
	"os"
	"time"

	"goji.io"

	"golang.org/x/net/context"

	"github.com/derekdowling/go-stdlogger"
	"github.com/zenazn/goji/web/mutil"
)

const (
	// FastResponse is anything under this duration
	FastResponse = 500 * time.Millisecond
	// AcceptableResponse is anything under this duration
	AcceptableResponse = 5 * time.Second
)

var logger std.Logger = log.New(os.Stderr, "", log.LstdFlags)

// SetLogger allows you to use your own logging solution
func SetLogger(newLogger std.Logger) {
	logger = newLogger
}

// Middleware logs the start and end of each request, along
// with some useful data about what was requested, what the response status was,
// and how long it took to return. When standard output is a TTY, Logger will
// print in color, otherwise it will print in black and white.
//
// Logger has been designed explicitly to be good enough for use in small
// applications and for people just getting started with Goji. It is expected
// that applications will eventually outgrow this middleware and replace it with
// a custom request logger, such as one that produces machine-parseable output,
// outputs logs to a different service (e.g., syslog), or formats lines like
// those printed elsewhere in the application.
func Middleware(next goji.Handler) goji.Handler {
	middleware := func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		printRequest(r)

		// WrapWriter lets us peek at ResponseWriter outputs
		lw := mutil.WrapWriter(w)

		startTime := time.Now()
		next.ServeHTTPC(ctx, lw, r)

		if lw.Status() == 0 {
			lw.WriteHeader(http.StatusOK)
		}

		finishTime := time.Now()

		printResponse(lw, finishTime.Sub(startTime))
	}

	return goji.HandlerFunc(middleware)
}

func printRequest(r *http.Request) {
	var buf bytes.Buffer

	buf.WriteString("Serving ")
	colorWrite(&buf, bMagenta, "%s ", r.Method)
	colorWrite(&buf, nBlue, "%q ", r.URL.String())
	buf.WriteString("from ")
	buf.WriteString(r.RemoteAddr)

	log.Print(buf.String())
}

func printResponse(w mutil.WriterProxy, delta time.Duration) {
	var buf bytes.Buffer

	buf.WriteString("Returning HTTP ")

	status := w.Status()
	if status < 200 {
		colorWrite(&buf, bBlue, "%03d", status)
	} else if status < 300 {
		colorWrite(&buf, bGreen, "%03d", status)
	} else if status < 400 {
		colorWrite(&buf, bCyan, "%03d", status)
	} else if status < 500 {
		colorWrite(&buf, bYellow, "%03d", status)
	} else {
		colorWrite(&buf, bRed, "%03d", status)
	}

	buf.WriteString(" in ")

	if delta < FastResponse {
		colorWrite(&buf, nGreen, "%s", delta.String())
	} else if delta < AcceptableResponse {
		colorWrite(&buf, nYellow, "%s", delta.String())
	} else {
		colorWrite(&buf, nRed, "%s", delta.String())
	}

	log.Print(buf.String())
}
