package bitculator

// Params is a bag of query parameters. Keys mirror the API's names (snake_case);
// values may be strings, numbers, bools, or slices (serialized as CSV). A nil
// Params (or a nil value) is fine and simply contributes nothing.
type Params = map[string]any
