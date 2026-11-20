package llmroute

import "errors"

var ErrAdminGateMissing = errors.New("llm route: route did not pass the admin gate")
