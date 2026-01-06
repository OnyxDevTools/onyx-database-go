package onyx

import (
	"bufio"
	"encoding/json"
	"io"
	"net/http"

	"github.com/OnyxDevTools/onyx-database-go/contract"
)

type streamIterator struct {
	resp    *http.Response
	scanner *bufio.Scanner
	current map[string]any
	err     error
}

func newStreamIterator(resp *http.Response) contract.Iterator {
	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)
	return &streamIterator{resp: resp, scanner: scanner}
}

func (s *streamIterator) Next() bool {
	if s.err != nil {
		return false
	}
	if !s.scanner.Scan() {
		if err := s.scanner.Err(); err != nil && err != io.EOF {
			s.err = err
		}
		return false
	}
	var m map[string]any
	if err := json.Unmarshal(s.scanner.Bytes(), &m); err != nil {
		s.err = err
		return false
	}
	s.current = m
	return true
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
