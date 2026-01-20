package impl

import (
	"bufio"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/OnyxDevTools/onyx-database-go/contract"
)

type aiChatStream struct {
	resp         *http.Response
	scanner      *bufio.Scanner
	current      contract.AIChatCompletionChunk
	err          error
	logger       logger
	logResponses bool
}

type logger interface {
	Printf(format string, v ...any)
}

func newAIChatStream(resp *http.Response, logger logger, logResponses bool) contract.AIChatStream {
	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 2*1024*1024)
	return &aiChatStream{resp: resp, scanner: scanner, logger: logger, logResponses: logResponses}
}

func (s *aiChatStream) Next() bool {
	if s.err != nil {
		return false
	}

	for {
		if !s.scanner.Scan() {
			if err := s.scanner.Err(); err != nil && err != io.EOF {
				s.err = err
			}
			return false
		}

		line := strings.TrimSpace(s.scanner.Text())
		if line == "" {
			continue
		}

		if !strings.HasPrefix(line, "data:") {
			continue
		}

		payload := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if payload == "[DONE]" {
			return false
		}

		var chunk contract.AIChatCompletionChunk
		if err := json.Unmarshal([]byte(payload), &chunk); err != nil {
			s.err = err
			return false
		}
		if s.logResponses && s.logger != nil {
			s.logger.Printf("[onyx] %s", payload)
		}
		s.current = chunk
		return true
	}
}

func (s *aiChatStream) Chunk() contract.AIChatCompletionChunk {
	return s.current
}

func (s *aiChatStream) Err() error {
	if s.err != nil {
		return s.err
	}
	return s.scanner.Err()
}

func (s *aiChatStream) Close() error {
	return s.resp.Body.Close()
}
