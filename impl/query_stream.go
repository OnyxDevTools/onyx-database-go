package impl

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/OnyxDevTools/onyx-database-go/contract"
	"github.com/OnyxDevTools/onyx-database-go/internal/httpclient"
	"github.com/OnyxDevTools/onyx-database-go/internal/msgpack"
)

type streamIterator struct {
	resp    *http.Response
	scanner *bufio.Scanner
	decoder *msgpack.Decoder
	current map[string]any
	err     error
}

func newStreamIterator(resp *http.Response) contract.Iterator {
	if httpclient.IsMessagePackResponse(resp) {
		return &streamIterator{resp: resp, decoder: msgpack.NewDecoder(resp.Body)}
	}
	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)
	return &streamIterator{resp: resp, scanner: scanner}
}

func (s *streamIterator) Next() bool {
	if s.err != nil {
		return false
	}
	if s.decoder != nil {
		return s.nextMessagePack()
	}
	for {
		if !s.scanner.Scan() {
			if err := s.scanner.Err(); err != nil && err != io.EOF {
				s.err = err
			}
			return false
		}
		raw := bytes.TrimSpace(s.scanner.Bytes())
		if len(raw) == 0 {
			continue
		}

		var m map[string]any
		if err := json.Unmarshal(raw, &m); err != nil {
			s.err = err
			return false
		}
		s.current = m
		return true
	}
}

func (s *streamIterator) nextMessagePack() bool {
	for {
		var value any
		if err := s.decoder.Decode(&value); err != nil {
			if !errors.Is(err, io.EOF) {
				s.err = err
			}
			return false
		}
		// The server may send a nil frame to make proxies flush headers before
		// the first entity is available.
		if value == nil {
			continue
		}
		row, ok := value.(map[string]any)
		if !ok {
			s.err = fmt.Errorf("msgpack entity stream contained %T; expected an object", value)
			return false
		}
		s.current = row
		return true
	}
}

func (s *streamIterator) Value() map[string]any {
	return s.current
}

func (s *streamIterator) Err() error {
	return s.err
}

func (s *streamIterator) Close() error {
	return s.resp.Body.Close()
}
