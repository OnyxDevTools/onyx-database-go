package commands

import "encoding/json"

// jsonMarshalIndent allows tests to inject failures around JSON encoding.
var jsonMarshalIndent = json.MarshalIndent
