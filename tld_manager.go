package tldwhois

import (
	"bufio"
	"context"
	"errors"
	"io"
	"net/http"
	"sync/atomic"
	"time"
)

var (
	ianaAddr = "whois.iana.org:43"

	tldsURL = "https://data.iana.org/TLD/tlds-alpha-by-domain.txt"

	ErrNoServerFound = errors.New("no whois server found")

	ErrTimeout = errors.New("parse tlds timeout")

	defaultTimeout = time.Second * 10
	defaultLimit   = 100
)

type tldManager struct {
	ctx     context.Context
	cancel  context.CancelFunc
	url     string
	c       chan string
	domains uint64
	errc    chan error
	stopc   chan struct{}
	donec   chan struct{}
	client  *http.Client
	limit   int
}

func newTldManager(url string) *tldManager {
	return newTldManagerWithLimitAndTimeout(url, defaultLimit, defaultTimeout)
}

func newTldManagerWithLimitAndTimeout(url string, limit int, timeout time.Duration) *tldManager {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)

	m := &tldManager{
		ctx:    ctx,
		cancel: cancel,
		url:    url,
		c:      make(chan string),
		errc:   make(chan error, 1),
		stopc:  make(chan struct{}),
		donec:  make(chan struct{}),
		client: &http.Client{},
		limit:  limit,
	}

	go m.run()

	return m
}

func (m *tldManager) inc() {
	atomic.AddUint64(&m.domains, 1)
}

func (m *tldManager) load() uint64 {
	return atomic.LoadUint64(&m.domains)
}

func (m *tldManager) run() {
	defer func() {
		close(m.errc)
		close(m.donec)
		close(m.c)
	}()
	resp, err := m.client.Get(m.url)
	if err != nil {
		m.errc <- err
		return
	}

	err = m.parseResponse(resp.Body)
	if err != nil {
		m.errc <- err
		return
	}
}

func (m *tldManager) parseResponse(r io.ReadCloser) error {
	defer r.Close()

	sc := bufio.NewScanner(r)
	for sc.Scan() {
		if int(m.load()) >= m.limit {
			return nil
		}

		line := sc.Text()
		if len(line) <= 0 {
			break
		}
		if line[0] == '#' {
			continue
		}
		select {
		case <-m.ctx.Done():
			return ErrTimeout
		case <-m.stopc:
			return nil
		case m.c <- line:
			m.inc()
		}
	}
	if err := sc.Err(); err != nil {
		return err
	}

	return nil
}

func (m *tldManager) stop() {
	close(m.stopc)
}

func (m *tldManager) TldC() <-chan string {
	return m.c
}

func (m *tldManager) Err() error {
	return <-m.errc
}
