package http_test

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	adaptershttp "agenticlens/internal/adapters/http"
)

// shortFrameDeadline caps deadlines the handler arms. Writing and flushing
// still use the actual HTTP writer and TCP socket, including zero deadlines.
type shortFrameDeadline struct {
	http.ResponseWriter
	budget      time.Duration
	largeWrite  chan struct{}
	signalWrite sync.Once
}

func (w *shortFrameDeadline) Unwrap() http.ResponseWriter { return w.ResponseWriter }

func (w *shortFrameDeadline) Flush() { _ = w.FlushError() }

func (w *shortFrameDeadline) FlushError() error {
	return http.NewResponseController(w.ResponseWriter).Flush()
}

func (w *shortFrameDeadline) SetWriteDeadline(deadline time.Time) error {
	shortened := time.Now().Add(w.budget)
	if !deadline.IsZero() && deadline.After(shortened) {
		deadline = shortened
	}
	return http.NewResponseController(w.ResponseWriter).SetWriteDeadline(deadline)
}

func (w *shortFrameDeadline) Write(data []byte) (int, error) {
	if len(data) > 512<<10 && bytes.HasPrefix(data, []byte("event: trace_update\n")) {
		w.signalWrite.Do(func() { close(w.largeWrite) })
	}
	return w.ResponseWriter.Write(data)
}

func TestLiveStreamWriteDeadlineReleasesStalledSocket(t *testing.T) {
	const frameBudget = 200 * time.Millisecond
	writer, service := liveStreamSQLiteFixture(t)
	largeWrite, handlerDone := make(chan struct{}), make(chan struct{})
	stream := adaptershttp.LiveStreamHandler(service)
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer close(handlerDone)
		stream.ServeHTTP(&shortFrameDeadline{ResponseWriter: w, budget: frameBudget, largeWrite: largeWrite}, r)
	}))
	server.Config.WriteTimeout = frameBudget
	server.Config.ConnState = func(conn net.Conn, state http.ConnState) {
		if state == http.StateNew {
			if err := conn.(*net.TCPConn).SetWriteBuffer(1024); err != nil {
				t.Errorf("limit TCP send buffer: %v", err)
			}
		}
	}
	server.Start()
	defer server.Close()

	conn, err := net.DialTimeout("tcp", server.Listener.Addr().String(), 3*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	if err := conn.(*net.TCPConn).SetReadBuffer(1024); err != nil {
		t.Fatal(err)
	}
	if err := conn.SetDeadline(time.Now().Add(3 * time.Second)); err != nil {
		t.Fatal(err)
	}
	if _, err := fmt.Fprint(conn, "GET / HTTP/1.1\r\nHost: localhost\r\n\r\n"); err != nil {
		t.Fatal(err)
	}
	response, err := http.ReadResponse(bufio.NewReader(conn), nil)
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != http.StatusOK {
		t.Fatalf("opening live stream returned %s", response.Status)
	}
	scanner := bufio.NewScanner(response.Body)
	if !scanner.Scan() || scanner.Text() != "event: session_tree" || !scanner.Scan() || !strings.HasPrefix(scanner.Text(), "data: ") {
		t.Fatalf("initial live snapshot unavailable: %q (%v)", scanner.Text(), scanner.Err())
	}

	// A healthy idle connection may wait longer than its frame budget.
	select {
	case <-handlerDone:
		t.Fatal("live stream ended while waiting for a new frame")
	case <-time.After(2 * frameBudget):
	}

	now := time.Now().UnixMilli()
	message := `{"role":"assistant","content":"` + strings.Repeat("x", 1<<20) + `"}`
	if _, err := writer.Exec("INSERT INTO session VALUES ('large', NULL, 'coder', 'fixture', 'Large trace', ?, ?)", now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := writer.Exec("INSERT INTO message VALUES ('turn', 'large', ?, ?)", now, message); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := service.Start(ctx); err != nil {
		t.Fatal(err)
	}
	select {
	case <-largeWrite:
	case <-handlerDone:
		t.Fatal("idle live stream could not write the next trace frame")
	case <-time.After(3 * time.Second):
		t.Fatal("SQLite trace did not reach the live stream")
	}
	// Keep the socket open without draining its receive buffer. The handler
	// must return and unsubscribe without a client close or a resumed write.
	select {
	case <-handlerDone:
	case <-time.After(2 * time.Second):
		t.Fatal("live handler remained blocked writing to a stalled client")
	}
}
