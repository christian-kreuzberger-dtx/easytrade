package price

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"dynatrace.com/easytrade/pricing-service/utils"
	"github.com/gin-gonic/gin"
)

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

func TestParseRecords(t *testing.T) {
	tests := []struct {
		raw     string
		want    int
		wantErr bool
	}{
		{"", 100, false},
		{"5", 5, false},
		{"1", 1, false},
		{"1000", 1000, false},
		{"1001", 1000, false},
		{"99999999", 1000, false},
		{"0", 0, true},
		{"-5", 0, true},
		{"abc", 0, true},
		{"1.5", 0, true},
		{"99999999999999999999", 0, true},
	}
	for _, tt := range tests {
		got, err := parseRecords(tt.raw)
		if (err != nil) != tt.wantErr || got != tt.want {
			t.Errorf("parseRecords(%q) = %d, %v; want %d, error=%v", tt.raw, got, err, tt.want, tt.wantErr)
		}
	}
}

func TestGetPricingDataForInstrumentRejectsInvalidRecords(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/v1/prices/instrument/:instrumentId", GetPricingDataForInstrument)

	for _, records := range []string{"abc", "0", "-1"} {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/prices/instrument/1?records="+records, nil)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("records=%s: got status %d, want 400", records, w.Code)
		}
	}
}
