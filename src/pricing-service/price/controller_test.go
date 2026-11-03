package price

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"dynatrace.com/easytrade/pricing-service/utils"
	"github.com/gin-gonic/gin"
)

func TestParseRecords(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		present bool
		want    int
		wantErr bool
	}{
		{"missing uses default", "", false, 100, false},
		{"one", "1", true, 1, false},
		{"max boundary", "1000", true, 1000, false},
		{"above max is capped", "1001", true, 1000, false},
		{"frontend value is capped", "1440", true, 1000, false},
		{"zero", "0", true, 0, true},
		{"negative", "-5", true, 0, true},
		{"non-numeric", "abc", true, 0, true},
		{"empty", "", true, 0, true},
		{"decimal", "1.5", true, 0, true},
		{"leading space", " 5", true, 0, true},
		{"overflow", "99999999999999999999", true, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseRecords(tt.raw, tt.present)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseRecords(%q, %v) error = %v, wantErr %v", tt.raw, tt.present, err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("parseRecords(%q, %v) = %d, want %d", tt.raw, tt.present, got, tt.want)
			}
		})
	}
}

func TestGetPricingDataForInstrumentInvalidRecords(t *testing.T) {
	tests := []struct {
		name   string
		accept string
		want   string
	}{
		{"json", gin.MIMEJSON, `"error"`},
		{"xml", gin.MIMEXML, `<error>`},
	}

	gin.SetMode(gin.TestMode)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(rec)
			ctx.Params = gin.Params{{Key: "instrumentId", Value: "1"}}
			ctx.Request = httptest.NewRequest(http.MethodGet, "/v1/prices/instrument/1?records=abc", nil)
			ctx.Request.Header.Set("Accept", tt.accept)

			GetPricingDataForInstrument(ctx)

			if rec.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
			}
			if !strings.Contains(rec.Body.String(), tt.want) {
				t.Errorf("body %q does not contain %q", rec.Body.String(), tt.want)
			}
		})
	}
}

func TestPrepareCSV(t *testing.T) {
	priceList := []price{
		{
			Id:           123,
			InstrumentId: 456,
			Timestamp:    "2001-01-01T01:01:01",
			Open:         10.01,
			High:         10.01,
			Low:          10.01,
			Close:        10.01,
		},
		{
			Id:           222,
			InstrumentId: 555,
			Timestamp:    "2002-02-02T02:02:02",
			Open:         20.02,
			High:         20.02,
			Low:          20.02,
			Close:        20.02,
		},
	}
	expected := "date, open, high, low, close, volume\n2001-01-01T01:01:01,10.010000,10.010000,10.010000,10.010000,200\n2002-02-02T02:02:02,20.020000,20.020000,20.020000,20.020000,200\n"
	csv := prepareCSV(priceList, utils.FakeIntProvider{})

	if csv != expected {
		t.Errorf("The CSV does not generate correctly! Got: '%s' while expecting '%s'", csv, expected)
	}
}
