package impl

import "github.com/OnyxDevTools/onyx-database-go/contract"

// toQueryResults converts generic data into contract.QueryResults.
func toQueryResults(items []map[string]any) contract.QueryResults {
	if items == nil {
		return contract.QueryResults{}
	}
	return contract.QueryResults(items)
}
