package price

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"dynatrace.com/easytrade/pricing-service/services"
	"dynatrace.com/easytrade/pricing-service/utils"
	"github.com/gin-gonic/gin"

	log "github.com/sirupsen/logrus"
)

const (
	defaultRecords = 100
	maxRecords     = 1000
)

type errorResult struct {
	Error string `json:"error" xml:"error"`
}

// @Summary		Get instrument prices
// @Description	Get current price of each instrument
// @Tags			Pricing-service
// @Accept			*/*
// @Produce		json
// @Produce		application/xml
// @Success		200	{object}	price.pricesResult
// @Router			/v1/prices/latest [get]
func GetCurrentPrices(ctx *gin.Context) {
	log.Info("Getting current prices")

	var priceList []price

	services.DB.Where("Timestamp = (?)", services.DB.Table("Pricing").Select("max(Timestamp)")).Find(&priceList)

	negotiateResponse(ctx, http.StatusOK, &pricesResult{
		Results: priceList,
	})
	services.SendDataToRabbitQueue(prepareCSV(priceList, utils.RandomIntProvider{}))
}

// @Summary		Get instrument price
// @Description	Get last price of instrument
// @Tags			Pricing-service
// @Accept			*/*
// @Produce		json
// @Produce		application/xml
// @Success		200	{object}	price.price
// @Router			/v1/prices/last [get]
func GetLastPrice(ctx *gin.Context) {
	log.Info("Getting last price")

	var lastPrice price

	services.DB.Last(&lastPrice)

	negotiateResponse(ctx, http.StatusOK, &lastPrice)
	services.SendDataToRabbitQueue(prepareCSV([]price{lastPrice}, utils.RandomIntProvider{}))
}

// @Summary		Get prices of a particular instrument
// @Description	Get specific number of records of particular instrument
// @Tags			Pricing-service
// @Accept			*/*
// @Produce		json
// @Produce		application/xml
// @Success		200	{object}	price.pricesResult
// @Failure		400	{object}	price.errorResult
// @Router			/v1/prices/instrument/{instrumentId} [get]
// @Param			instrumentId	path	int	true	"Instrument id"
// @Param			records			query	int	false	"Number of records (1-1000, default 100; values above 1000 are capped)"
func GetPricingDataForInstrument(ctx *gin.Context) {
	instrumentId := ctx.Param("instrumentId")
	rawRecords, present := ctx.GetQuery("records")
	records, err := parseRecords(rawRecords, present)
	if err != nil {
		negotiateResponse(ctx, http.StatusBadRequest, &errorResult{Error: err.Error()})
		return
	}

	log.WithFields(log.Fields{
		"instrumentId": instrumentId,
		"records":      records,
	}).Info("Getting pricing data for instrument")

	var priceList []price

	services.DB.Table("Pricing").Where("instrumentId = ?", instrumentId).Order("Timestamp desc").Limit(records).Scan(&priceList)

	negotiateResponse(ctx, http.StatusOK, &pricesResult{
		Results: priceList,
	})
	services.SendDataToRabbitQueue(prepareCSV(priceList, utils.RandomIntProvider{}))
}

// parseRecords protects the DB from invalid limits and oversized queries.
func parseRecords(raw string, present bool) (int, error) {
	if !present {
		return defaultRecords, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n <= 0 {
		return 0, fmt.Errorf("invalid 'records' parameter %q: must be an integer between 1 and %d", raw, maxRecords)
	}
	return min(n, maxRecords), nil
}

func prepareCSV(priceList []price, provider utils.IntProvider) string {
	var stringBuilder strings.Builder
	stringBuilder.WriteString("date, open, high, low, close, volume\n")

	for _, item := range priceList {
		stringBuilder.WriteString(item.toCSV(provider.Intn(100) + 100))
	}

	return stringBuilder.String()
}

func negotiateResponse(ctx *gin.Context, status int, data any) {
	ctx.Negotiate(status, gin.Negotiate{
		Offered: []string{gin.MIMEJSON, gin.MIMEXML},
		Data:    data,
	})
}
