package offer

import (
	"reflect"
	"testing"
)

func TestParseResponse(t *testing.T) {
	tests := []struct {
		name      string
		body      string
		mimeType  string
		want      *Offer
		wantError bool
	}{
		{
			name:     "JSON",
			body:     `{"platform":"web","quoteFor":"user"}`,
			mimeType: jsonMimeType,
			want:     &Offer{Platform: "web", QuoteFor: "user"},
		},
		{
			name:     "XML",
			body:     `<offer><platform>web</platform><quoteFor>user</quoteFor></offer>`,
			mimeType: xmlMimeType,
			want:     &Offer{Platform: "web", QuoteFor: "user"},
		},
		{
			name:      "invalid JSON",
			body:      `{`,
			mimeType:  jsonMimeType,
			wantError: true,
		},
		{
			name:      "invalid XML",
			body:      `<offer>`,
			mimeType:  xmlMimeType,
			wantError: true,
		},
		{
			name:      "unsupported MIME type",
			mimeType:  "text/plain",
			want:      nil,
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseResponse([]byte(tt.body), tt.mimeType)
			if (err != nil) != tt.wantError {
				t.Fatalf("parseResponse() error = %v, wantError %v", err, tt.wantError)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseResponse() = %#v, want %#v", got, tt.want)
			}
		})
	}
}
