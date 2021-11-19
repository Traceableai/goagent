//go:build linux && traceable_filter
// +build linux,traceable_filter

package traceable

const (
	grpcContentType    = "application/grpc"
	grpcContentTypeLen = 16 // len(grpcContentType)
)

// isGRPC determines whether a metadata set belongs to http or no
func isGRPC(h map[string][]string) bool {
	// metadata always does the lookup in lowercase
	// see https://pkg.go.dev/google.golang.org/grpc@v1.42.0/metadata#MD.Get
	contentType, ok := h["content-type"]
	if !ok || len(contentType) == 0 || len(contentType[0]) < grpcContentTypeLen {
		return false
	}

	return contentType[0][:grpcContentTypeLen] == grpcContentType
}
