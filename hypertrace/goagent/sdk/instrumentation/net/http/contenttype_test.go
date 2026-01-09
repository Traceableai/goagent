package http

import (
	"net/http"
	"testing"

	internalconfig "github.com/Traceableai/goagent/hypertrace/goagent/sdk/internal/config"
	config "github.com/hypertrace/agent-config/gen/go/v1"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/stretchr/testify/assert"
)

func TestRecordingDecisionReturnsFalseOnNoContentType(t *testing.T) {
	assert.Equal(t, false, ShouldRecordBodyOfContentType(&headerMapAccessor{http.Header{"A": []string{"B"}}}))
}

func TestRecordingDecisionWildcardContentType(t *testing.T) {
	t.Cleanup(internalconfig.ResetConfig)
	cfg := config.Load()
	cfg.DataCapture.AllowedContentTypes = []*wrapperspb.StringValue{wrapperspb.String("*")}
	internalconfig.ResetConfig()
	internalconfig.InitConfig(cfg)

	tests := []struct {
		name           string
		headerAccessor HeaderAccessor
	}{
		{
			name:           "text/plain",
			headerAccessor: &headerMapAccessor{http.Header{"Content-Type": []string{"text/plain"}}},
		},
		{
			name:           "application/json",
			headerAccessor: &headerMapAccessor{http.Header{"Content-Type": []string{"application/json"}}},
		},
		{
			name:           "no value present",
			headerAccessor: &headerMapAccessor{http.Header{"A": []string{"B"}}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.True(t, ShouldRecordBodyOfContentType(tt.headerAccessor))
		})
	}
}

func TestRecordingDecisionSuccessOnHeaderSet(t *testing.T) {
	internalconfig.ResetConfig()
	tCases := []struct {
		contentType  string
		shouldRecord bool
	}{
		{"text/plain", false},
		{"application/json", true},
		{"Application/JSON", true},
		{"application/json", true},
		{"application/json; charset=utf-8", true},
		{"application/x-www-form-urlencoded", true},
		{"application/vnd.api+json", true},
		{"application/grpc+json", true},
	}

	for _, tCase := range tCases {
		h := http.Header{}
		h.Set("Content-Type", tCase.contentType)
		assert.Equal(t, tCase.shouldRecord, ShouldRecordBodyOfContentType(&headerMapAccessor{h}))
	}
}

func TestRecordingDecisionSuccessOnHeaderAdd(t *testing.T) {
	internalconfig.ResetConfig()
	tCases := []struct {
		contentTypes []string
		shouldRecord bool
	}{
		{[]string{"text/plain"}, false},
		{[]string{"application/json"}, true},
		{[]string{"application/json", "charset=utf-8"}, true},
		{[]string{"application/json; charset=utf-8"}, true},
		{[]string{"application/x-www-form-urlencoded"}, true},
		{[]string{"charset=utf-8", "application/json"}, true},
		{[]string{"charset=utf-8", "application/vnd.api+json"}, true},
	}

	for _, tCase := range tCases {
		h := http.Header{}
		for _, header := range tCase.contentTypes {
			h.Add("Content-Type", header)
		}
		assert.Equal(t, tCase.shouldRecord, ShouldRecordBodyOfContentType(&headerMapAccessor{h}))
	}
}

func TestXMLRecordingDecisionSuccessOnHeaderAdd(t *testing.T) {
	cfg := internalconfig.GetConfig()
	cfg.DataCapture.AllowedContentTypes = []*wrapperspb.StringValue{wrapperspb.String("xml")}

	tCases := []struct {
		contentTypes []string
		shouldRecord bool
	}{
		{[]string{"text/xml"}, true},
		{[]string{"application/xml"}, true},
		{[]string{"image/svg+xml"}, true},
		{[]string{"application/xhtml+xml"}, true},
		{[]string{"text/plain"}, false},
	}

	for _, tCase := range tCases {
		h := http.Header{}
		for _, header := range tCase.contentTypes {
			h.Add("Content-Type", header)
		}
		assert.Equal(t, tCase.shouldRecord, ShouldRecordBodyOfContentType(&headerMapAccessor{h}))
	}
	internalconfig.ResetConfig()
}

func TestHasMultiPartFormDataContentTypeHeader(t *testing.T) {
	tCases := []struct {
		contentType         string
		isMultiPartFormData bool
	}{
		{"text/plain", false},
		{"application/json", false},
		{"multipart/form-data", true},
		{"multipart/mixed", false},
	}

	for _, tCase := range tCases {
		h := http.Header{}
		h.Set("Content-Type", tCase.contentType)
		assert.Equal(t, tCase.isMultiPartFormData, HasMultiPartFormDataContentTypeHeader(&headerMapAccessor{h}))
	}
}
