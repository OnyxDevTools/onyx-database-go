package resolver

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/OnyxDevTools/onyx-database-go/contract"
	"github.com/OnyxDevTools/onyx-database-go/internal/httpclient"
)

var (
	resolveDatabaseIDFromAPIKeyFn = resolveDatabaseIDFromAPIKey
	lookupDatabaseIDFromAPIKeyFn  = lookupDatabaseIDFromAPIKey

	uuidPattern    = regexp.MustCompile(`(?i)[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`)
	dbTokenPattern = regexp.MustCompile(`(?i)\bdb[_-][a-z0-9][a-z0-9_-]{2,}\b`)

	databaseIDResolvePaths = []string{
		"/database/resolve",
		"/database",
	}
)

type databaseIDResolveResponse struct {
	DatabaseID string `json:"databaseId"`
	ID         string `json:"id"`
	Database   struct {
		DatabaseID string `json:"databaseId"`
		ID         string `json:"id"`
	} `json:"database"`
	Databases []struct {
		DatabaseID string `json:"databaseId"`
		ID         string `json:"id"`
	} `json:"databases"`
	Data []struct {
		DatabaseID string `json:"databaseId"`
		ID         string `json:"id"`
	} `json:"data"`
}

func (r databaseIDResolveResponse) firstID() string {
	if strings.TrimSpace(r.DatabaseID) != "" {
		return strings.TrimSpace(r.DatabaseID)
	}
	if strings.TrimSpace(r.ID) != "" {
		return strings.TrimSpace(r.ID)
	}
	if strings.TrimSpace(r.Database.DatabaseID) != "" {
		return strings.TrimSpace(r.Database.DatabaseID)
	}
	if strings.TrimSpace(r.Database.ID) != "" {
		return strings.TrimSpace(r.Database.ID)
	}
	if len(r.Databases) > 0 {
		if strings.TrimSpace(r.Databases[0].DatabaseID) != "" {
			return strings.TrimSpace(r.Databases[0].DatabaseID)
		}
		if strings.TrimSpace(r.Databases[0].ID) != "" {
			return strings.TrimSpace(r.Databases[0].ID)
		}
	}
	if len(r.Data) > 0 {
		if strings.TrimSpace(r.Data[0].DatabaseID) != "" {
			return strings.TrimSpace(r.Data[0].DatabaseID)
		}
		if strings.TrimSpace(r.Data[0].ID) != "" {
			return strings.TrimSpace(r.Data[0].ID)
		}
	}
	return ""
}

func resolveDatabaseIDFromAPIKey(ctx context.Context, baseURL, apiKey, apiSecret string) (string, error) {
	if dbID := extractDatabaseIDFromAPIKey(apiKey); dbID != "" {
		return dbID, nil
	}

	if strings.TrimSpace(apiKey) == "" || strings.TrimSpace(apiSecret) == "" {
		return "", nil
	}
	if strings.TrimSpace(baseURL) == "" {
		return "", errors.New("resolve database id: missing base url")
	}

	return lookupDatabaseIDFromAPIKeyFn(ctx, baseURL, apiKey, apiSecret)
}

func extractDatabaseIDFromAPIKey(apiKey string) string {
	key := strings.TrimSpace(apiKey)
	if key == "" {
		return ""
	}
	if match := uuidPattern.FindString(key); match != "" {
		return match
	}
	if match := dbTokenPattern.FindString(key); match != "" {
		return match
	}
	return ""
}

func lookupDatabaseIDFromAPIKey(ctx context.Context, baseURL, apiKey, apiSecret string) (string, error) {
	client := httpclient.New(baseURL, nil, httpclient.Options{
		Signer: httpclient.Signer{APIKey: apiKey, APISecret: apiSecret},
	})

	base := strings.TrimRight(baseURL, "/")
	var lastErr error
	for _, path := range databaseIDResolvePaths {
		if debugEnabled() {
			log.Printf("onyx resolver: resolve database id GET %s%s", base, path)
		}
		var resp databaseIDResolveResponse
		if err := client.DoJSON(ctx, http.MethodGet, path, nil, &resp); err != nil {
			if debugEnabled() {
				log.Printf("onyx resolver: resolve database id failed for %s%s: %v", base, path, err)
			}
			if shouldTryNextResolvePath(err) {
				lastErr = err
				continue
			}
			return "", err
		}
		if dbID := resp.firstID(); dbID != "" {
			return dbID, nil
		}
		return "", fmt.Errorf("resolve database id: response missing databaseId")
	}

	if lastErr != nil {
		return "", lastErr
	}
	return "", errors.New("resolve database id: no resolve endpoints succeeded")
}

func shouldTryNextResolvePath(err error) bool {
	var cerr *contract.Error
	if !errors.As(err, &cerr) {
		return false
	}
	status, ok := cerr.Meta["status"].(int)
	if !ok {
		if statusFloat, ok := cerr.Meta["status"].(float64); ok {
			status = int(statusFloat)
		}
	}
	return status == http.StatusNotFound || status == http.StatusMethodNotAllowed
}

func debugEnabled() bool {
	return readConfigEnv("ONYX_DEBUG") == "true"
}
