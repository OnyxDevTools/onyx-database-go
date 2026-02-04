package resolver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	defaultAIBaseURL       = "https://ai.onyx.dev"
	defaultDatabaseBaseURL = "https://api.onyx.dev"
)

var (
	pathAbs            = filepath.Abs
	evalSymlinks       = filepath.EvalSymlinks
	readConfigEnv      = func(key string) string { return os.Getenv(key) }
	resolveFromFilesFn = resolveFromFiles
)

// Config captures user-provided configuration values used during resolution.
type Config struct {
	DatabaseID      string
	DatabaseBaseURL string
	APIKey          string
	APISecret       string
	AIBaseURL       string
	CacheTTL        time.Duration
	ConfigPath      string
	LogRequests     bool
	LogResponses    bool
	Partition       string
}

// ResolvedConfig represents a fully-hydrated configuration ready for use.
type ResolvedConfig struct {
	DatabaseID      string
	DatabaseBaseURL string
	APIKey          string
	APISecret       string
	AIBaseURL       string
	CacheTTL        time.Duration
	LogRequests     bool
	LogResponses    bool
	Partition       string
}

// Source identifies how a particular field was populated.
type Source string

const (
	SourceNone     Source = ""
	SourceExplicit Source = "explicit"
	SourceEnv      Source = "env"
	SourceFile     Source = "file"
	SourceDerived  Source = "derived"
	SourceDefault  Source = "default"
)

// Meta contains debug information about configuration resolution.
type Meta struct {
	Sources  FieldSources
	FilePath string
}

// FieldSources holds the origin for each resolved value.
type FieldSources struct {
	DatabaseID      Source
	DatabaseBaseURL Source
	APIKey          Source
	APISecret       Source
	AIBaseURL       Source
}

// Resolve merges explicit values, environment variables, and configuration files.
func Resolve(ctx context.Context, partial Config) (ResolvedConfig, Meta, error) {
	key := cacheKey(partial)
	if cfg, meta, ok := defaultCache.get(key); ok {
		return cfg, meta, nil
	}

	resolved := ResolvedConfig{}
	meta := Meta{Sources: FieldSources{}}

	apply := func(val string, target *string, src *Source, sourceVal Source) {
		if val == "" {
			return
		}
		*target = val
		*src = sourceVal
	}

	// Explicit configuration wins above all else.
	apply(partial.DatabaseID, &resolved.DatabaseID, &meta.Sources.DatabaseID, SourceExplicit)
	apply(partial.DatabaseBaseURL, &resolved.DatabaseBaseURL, &meta.Sources.DatabaseBaseURL, SourceExplicit)
	apply(partial.APIKey, &resolved.APIKey, &meta.Sources.APIKey, SourceExplicit)
	apply(partial.APISecret, &resolved.APISecret, &meta.Sources.APISecret, SourceExplicit)
	apply(partial.AIBaseURL, &resolved.AIBaseURL, &meta.Sources.AIBaseURL, SourceExplicit)

	// Environment variables fill missing values.
	envVars := map[string]struct {
		ptr *string
		src *Source
		key string
	}{
		"id":     {&resolved.DatabaseID, &meta.Sources.DatabaseID, "ONYX_DATABASE_ID"},
		"base":   {&resolved.DatabaseBaseURL, &meta.Sources.DatabaseBaseURL, "ONYX_DATABASE_BASE_URL"},
		"apiKey": {&resolved.APIKey, &meta.Sources.APIKey, "ONYX_DATABASE_API_KEY"},
		"secret": {&resolved.APISecret, &meta.Sources.APISecret, "ONYX_DATABASE_API_SECRET"},
		"aiBase": {&resolved.AIBaseURL, &meta.Sources.AIBaseURL, "ONYX_AI_BASE_URL"},
	}

	for _, spec := range envVars {
		if *spec.ptr == "" {
			if v := strings.TrimSpace(os.Getenv(spec.key)); v != "" {
				apply(v, spec.ptr, spec.src, SourceEnv)
			}
		}
	}

	// File-based config is lowest precedence.
	filePath, err := resolveFromFilesFn(ctx, partial, &resolved, &meta)
	if err != nil {
		return ResolvedConfig{}, Meta{}, err
	}
	if filePath != "" {
		if abs, err := pathAbs(filePath); err == nil {
			if resolved, err := evalSymlinks(abs); err == nil {
				meta.FilePath = resolved
			} else {
				meta.FilePath = abs
			}
		} else {
			meta.FilePath = filePath
		}
	} else {
		meta.FilePath = filePath
	}

	if strings.TrimSpace(resolved.DatabaseBaseURL) == "" {
		resolved.DatabaseBaseURL = defaultDatabaseBaseURL
		meta.Sources.DatabaseBaseURL = SourceDefault
	}

	if resolved.DatabaseID == "" && strings.TrimSpace(resolved.APIKey) != "" {
		dbID, err := resolveDatabaseIDFromAPIKeyFn(ctx, resolved.DatabaseBaseURL, resolved.APIKey, resolved.APISecret)
		if err != nil {
			return ResolvedConfig{}, Meta{}, err
		}
		if strings.TrimSpace(dbID) != "" {
			resolved.DatabaseID = strings.TrimSpace(dbID)
			meta.Sources.DatabaseID = SourceDerived
		}
	}

	if resolved.DatabaseID == "" || resolved.DatabaseBaseURL == "" || resolved.APIKey == "" || resolved.APISecret == "" {
		return ResolvedConfig{}, Meta{}, errors.New("missing required configuration values")
	}

	if strings.TrimSpace(resolved.AIBaseURL) == "" {
		resolved.AIBaseURL = defaultAIBaseURL
	}

	resolved.CacheTTL = partial.CacheTTL
	if resolved.CacheTTL <= 0 {
		resolved.CacheTTL = 5 * time.Minute
	}
	resolved.LogRequests = partial.LogRequests
	resolved.LogResponses = partial.LogResponses
	if strings.TrimSpace(partial.Partition) != "" {
		resolved.Partition = strings.TrimSpace(partial.Partition)
	}

	ttl := resolved.CacheTTL
	defaultCache.set(key, resolved, meta, ttl)
	if os.Getenv("ONYX_DEBUG") == "true" {
		log.Printf("onyx resolver: sources id=%s base=%s key=%s secret=%s file=%s partition=%v", meta.Sources.DatabaseID, meta.Sources.DatabaseBaseURL, meta.Sources.APIKey, meta.Sources.APISecret, meta.FilePath, resolved.Partition != "")
	}
	return resolved, meta, nil
}

func cacheKey(cfg Config) string {
	// A deterministic cache key based on explicit inputs only; other sources do not affect the key.
	b, _ := json.Marshal(cfg)
	return string(b)
}

type fileConfig struct {
	DatabaseID      string `json:"databaseId"`
	DatabaseBaseURL string `json:"databaseBaseUrl"`
	BaseURL         string `json:"baseUrl"`
	APIKey          string `json:"apiKey"`
	APISecret       string `json:"apiSecret"`
	Partition       string `json:"partition"`
	AIBaseURL       string `json:"aiBaseUrl"`
}

func resolveFromFiles(ctx context.Context, partial Config, resolved *ResolvedConfig, meta *Meta) (string, error) {
	path := partial.ConfigPath
	if path == "" {
		path = readConfigEnv("ONYX_CONFIG_PATH")
	}
	var chosenPath string

	candidates := []string{}
	if path != "" {
		candidates = append(candidates, path)
	} else {
		dbID := resolved.DatabaseID
		if dbID == "" {
			dbID = partial.DatabaseID
		}

		// Search order: config dir first, then project root variants.
		if dbID != "" {
			candidates = append(candidates, filepath.Join("config", fmt.Sprintf("onyx-database-%s.json", dbID)))
		}
		candidates = append(candidates, filepath.Join("config", "onyx-database.json"))
		if dbID != "" {
			candidates = append(candidates, filepath.Join("internal", "config", fmt.Sprintf("onyx-database-%s.json", dbID)))
		}
		candidates = append(candidates, filepath.Join("internal", "config", "onyx-database.json"))
		if dbID != "" {
			candidates = append(candidates, fmt.Sprintf("./onyx-database-%s.json", dbID))
		}
		candidates = append(candidates, "./onyx-database.json")

		if home, err := os.UserHomeDir(); err == nil {
			if dbID != "" {
				candidates = append(candidates, filepath.Join(home, ".onyx", fmt.Sprintf("onyx-database-%s.json", dbID)))
				candidates = append(candidates, filepath.Join(home, fmt.Sprintf("onyx-database-%s.json", dbID)))
			}
			candidates = append(candidates, filepath.Join(home, ".onyx", "onyx-database.json"))
			candidates = append(candidates, filepath.Join(home, "onyx-database.json"))
		}
	}

	for _, candidate := range candidates {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}

		data, err := os.ReadFile(candidate)
		if err != nil {
			continue
		}

		var fc fileConfig
		if err := json.Unmarshal(data, &fc); err != nil {
			continue
		}

		applied := false
		if resolved.DatabaseID == "" && fc.DatabaseID != "" {
			resolved.DatabaseID = fc.DatabaseID
			meta.Sources.DatabaseID = SourceFile
			applied = true
		}
		if resolved.DatabaseBaseURL == "" {
			switch {
			case fc.DatabaseBaseURL != "":
				resolved.DatabaseBaseURL = fc.DatabaseBaseURL
				meta.Sources.DatabaseBaseURL = SourceFile
				applied = true
			case fc.BaseURL != "":
				resolved.DatabaseBaseURL = fc.BaseURL
				meta.Sources.DatabaseBaseURL = SourceFile
				applied = true
			}
		}
		if resolved.APIKey == "" && fc.APIKey != "" {
			resolved.APIKey = fc.APIKey
			meta.Sources.APIKey = SourceFile
			applied = true
		}
		if resolved.APISecret == "" && fc.APISecret != "" {
			resolved.APISecret = fc.APISecret
			meta.Sources.APISecret = SourceFile
			applied = true
		}
		if resolved.AIBaseURL == "" && strings.TrimSpace(fc.AIBaseURL) != "" {
			resolved.AIBaseURL = strings.TrimSpace(fc.AIBaseURL)
			meta.Sources.AIBaseURL = SourceFile
			applied = true
		}
		if resolved.Partition == "" && strings.TrimSpace(fc.Partition) != "" {
			resolved.Partition = strings.TrimSpace(fc.Partition)
			applied = true
		}

		if applied {
			chosenPath = candidate
		}

		if resolved.DatabaseID != "" && resolved.DatabaseBaseURL != "" && resolved.APIKey != "" && resolved.APISecret != "" {
			break
		}
	}

	return chosenPath, nil
}
