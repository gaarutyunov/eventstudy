package main

import (
	"context"
	"flag"
	"fmt"
	sdk "github.com/TinkoffCreditSystems/invest-openapi-go-sdk"
	"github.com/gaarutyunov/eventstudy/internal/candle"
	"github.com/gaarutyunov/eventstudy/internal/instrument"
	"github.com/spf13/viper"
	"go.uber.org/ratelimit"
	"log"
	"time"
)

var ticker = flag.String("ticker", "TCS", "instrument ticker")
var benchmark = flag.String("benchmark", "MOEX", "benchmark market index")
var rf = flag.Float64("rf", 0.045, "risk free rate")
var from = flag.String("from", "2019-01-01", "starting date")
var to = flag.String("to", "2020-10-04", "ending date")
var event = flag.String("event", "2020-03-06", "event date")
var window = flag.Int("window", 50, "rolling window")
var period = flag.Int("period", 50, "period of estimation")
var del = flag.String("delimiter", ",", "delimiter for output file")
var output = flag.String("out", "out/returns.csv", "output file")

const dateLayout = "2006-01-02"

func main() {
	flag.Parse()

	fromDate, err := time.Parse(dateLayout, *from)

	if err != nil {
		log.Fatalf("error parsing `from` date: %v", err)
	}

	toDate, err := time.Parse(dateLayout, *to)

	if err != nil {
		log.Fatalf("error parsing `to` date: %v", err)
	}

	eventDate, err := time.Parse(dateLayout, *event)

	if err != nil {
		log.Fatalf("error parsing `event` date: %v", err)
	}

	var d [][]time.Time

	dates(fromDate, toDate, &d)

	viper.AddConfigPath(".")
	viper.SetConfigName("config")
	viper.SetConfigType("yml")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("error reading config: %v", err)
	}

	client := sdk.NewSandboxRestClient(viper.GetString("api.token"))

	i := &instrument.Instrument{
		Ticker:  ticker,
		Candles: &candle.Candles{},
	}

	rm := &instrument.Instrument{
		Ticker:  benchmark,
		Candles: &candle.Candles{},
	}

	rtl := ratelimit.New(2)

	err = fetchInstrument(i, d, client, rtl)
	if err != nil {
		log.Fatalf("error fetching data for %s", *i.Ticker)
	}

	err = fetchInstrument(rm, d, client, rtl)
	if err != nil {
		log.Fatalf("error fetching data for %s", *rm.Ticker)
	}

	rm.CalculateReturns()
	err = i.CalculateReturns().
		Capm(fromDate, eventDate.Add(-time.Hour * 24), rm).
		EstimateReturns(*window, *period, *rf).
		CalculateAbnormalReturns().
		ToCsv(*output, &[]rune(*del)[0])

	if err != nil {
		log.Fatal(err)
	}
}

func fetchInstrument(i *instrument.Instrument, d [][]time.Time, client *sdk.SandboxRestClient, rtl ratelimit.Limiter) error {
	rtl.Take()

	ctx, cancel := newCtx()

	t, e := client.InstrumentByTicker(ctx, *i.Ticker)

	if e != nil || len(t) == 0 {
		cancel()
		return fmt.Errorf("error getting ticker %s: %v", *i.Ticker, e)
	}

	for _, ft := range d {
		c, e := fetchCandles(rtl, ft, t, client)
		if e != nil {
			return e
		}
		i.Candles.C = append(i.Candles.C, c...)
	}

	return nil
}

func fetchCandles(rtl ratelimit.Limiter, ft []time.Time, t []sdk.Instrument, client *sdk.SandboxRestClient) ([]*candle.Candle, error) {
	rtl.Take()

	ctx, cancel := newCtx()

	c, e := client.Candles(ctx, ft[0], ft[1], sdk.CandleInterval1Day, t[0].FIGI)

	if e != nil || len(c) == 0 {
		cancel()
		return nil, fmt.Errorf("error getting candles from %s to %s for %s: %v", ft[0], ft[1], t[0].Ticker, e)
	}

	var candles []*candle.Candle

	for _, cdl := range c {
		candles = append(candles, &candle.Candle{
			Open:   cdl.OpenPrice,
			High:   cdl.HighPrice,
			Close:  cdl.ClosePrice,
			Low:    cdl.LowPrice,
			Volume: cdl.LowPrice,
			Date:   time.Date(cdl.TS.Year(), cdl.TS.Month(), cdl.TS.Day(), 0,0,0,0, cdl.TS.Location()),
		})
	}

	return candles, nil
}

func newCtx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), time.Minute * 2)
}

func dates(from, to time.Time, out *[][]time.Time) {
	if from.Add(time.Hour * 24 * 365).After(to) {
		*out = append(*out, []time.Time{from, to})
		return
	}

	nFrom := from.Add(time.Hour * 24 * 365)

	*out = append(*out, []time.Time{from, nFrom.Add(-time.Hour * 24)})

	dates(nFrom, to, out)

	return
}
