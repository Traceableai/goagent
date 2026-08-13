package result

type KeyValueString struct {
	Key   string
	Value string
}

type Decorations struct {
	RequestHeaderInjections   []KeyValueString
	RequestBodyModifications  string
	ResponseBodyModifications string
}

type FilterResult struct {
	Block              bool
	ResponseStatusCode int32
	ResponseMessage    string
	Decorations        *Decorations
	OutAttributes      map[string]string
}
