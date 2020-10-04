package candle

import "time"

type Candle struct {
	Open   float64
	High   float64
	Close  float64
	Low    float64
	Volume float64
	Date time.Time
}

type Candles struct {
	C []*Candle
}

func (c Candles) Len() int {
	return len(c.C)
}

func (c Candles) Less(i, j int) bool {
	return c.C[i].Date.Before(c.C[j].Date)
}

func (c Candles) Swap(i, j int) {
	c.C[i], c.C[j] = c.C[j], c.C[i]
}
