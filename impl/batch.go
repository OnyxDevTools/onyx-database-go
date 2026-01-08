package impl

import (
	"context"
	"net/http"
	"time"

	"github.com/OnyxDevTools/onyx-database-go/contract"
)

const defaultBatchSize = 500

var transientStatus = map[int]bool{
	http.StatusTooManyRequests:    true,
	http.StatusBadGateway:         true,
	http.StatusServiceUnavailable: true,
	http.StatusGatewayTimeout:     true,
}

func batchSave(ctx context.Context, c *client, table string, entities []any, batchSize int) error {
	if batchSize <= 0 {
		batchSize = defaultBatchSize
	}

	path := c.tablePath(table)

	for start := 0; start < len(entities); start += batchSize {
		end := start + batchSize
		if end > len(entities) {
			end = len(entities)
		}
		chunk := entities[start:end]
		// Match TS SDK: send the slice directly (not wrapped) so the API receives an array of entities.
		payload := chunk

		err := c.httpClient.DoJSON(ctx, http.MethodPut, path, payload, nil)
		if err == nil {
			continue
		}

		if cerr, ok := err.(*contract.Error); ok {
			if status, ok := cerr.Meta["status"].(int); ok && transientStatus[status] {
				// single retry after small backoff
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(50 * time.Millisecond):
				}
				if retryErr := c.httpClient.DoJSON(ctx, http.MethodPut, path, payload, nil); retryErr == nil {
					continue
				} else {
					return retryErr
				}
			}
		}
		return err
	}

	return nil
}
