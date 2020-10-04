package instrument

import (
	"fmt"
	"github.com/gaarutyunov/eventstudy/internal/candle"
	"github.com/gaarutyunov/eventstudy/internal/returns"
	"os"
	"path/filepath"
	"sort"
	"time"
)

type Instrument struct {
	Ticker  *string
	Candles *candle.Candles
	Returns *returns.Returns
}

type CAPM struct {
	Ticker  *string
	Returns *returns.Returns
	Market  *returns.Returns
	Rm      *Instrument
	I       *Instrument
	mrm     map[time.Time]*returns.Return
}

type EventStudy struct {
	C  *CAPM
	AR *returns.Returns
}

type Beta struct {
	Date time.Time
	B    float64
}

const (
	layout = "2006-01-02"
	defaultDel = ','
)

func (i *Instrument) CalculateReturns() *Instrument {
	r := make([]*returns.Return, len(i.Candles.C))

	for idx, c := range i.Candles.C {
		if idx == 0 {
			r[0] = &returns.Return{
				Date: c.Date,
				R:    0,
			}
			continue
		}

		ret := &returns.Return{
			Date: c.Date,
		}

		ret.R = (c.Close - i.Candles.C[idx-1].Close) / i.Candles.C[idx-1].Close

		r[idx] = ret
	}

	i.Returns = &returns.Returns{
		R: r,
	}

	return i
}

func (i *Instrument) Capm(from, to time.Time, rm *Instrument) *CAPM {
	c := &CAPM{
		Ticker:  i.Ticker,
		Returns: &returns.Returns{},
		Market:  &returns.Returns{},
		I:       i,
		Rm:      rm,
		mrm:     map[time.Time]*returns.Return{},
	}

	mi := map[time.Time]*returns.Return{}
	mrm := map[time.Time]*returns.Return{}
	dates := []time.Time{}

	for _, r := range i.Returns.R {
		mi[r.Date] = r
		dates = append(dates, r.Date)
	}
	for _, r := range rm.Returns.R {
		if _, ok := mi[r.Date]; ok {
			mrm[r.Date] = r
			c.Market.R = append(c.Market.R, r)
		}
	}

	var idx int
	for d, r := range mi {
		v, ok := mrm[d]

		if ok {
			c.mrm[v.Date] = v
		}

		idx++

		if r.Date.Before(from) || r.Date.After(to) {
			continue
		}

		c.Returns.R = append(c.Returns.R, r)
	}

	return c
}

func (c *CAPM) EstimateReturns(window, period int, rf float64) *CAPM {
	if !sort.IsSorted(c.Returns) {
		sort.Sort(c.Returns)
	}

	if !sort.IsSorted(c.Rm.Returns) {
		sort.Sort(c.Rm.Returns)
	}

	c.Rm.Returns.CalculateMovingAverage(window)

	c.Returns.CalculateMovingAverage(window).
		CalculateMovingVariance(window).
		CalculateMovingCovariance(window, c.Rm.Returns)

	for idx := 0; idx < period; idx++ {
		c.Returns.AppendMovingAverage(window, len(c.Returns.R)-1).
			AppendMovingVariance(len(c.Returns.R)-1).
			AppendMovingCovariance(window, c.Rm.Returns, len(c.Returns.R)-1)

		market := c.Market.R[len(c.Returns.R)]

		re := rf + c.Returns.Beta()*(market.R-rf)

		c.Returns.R = append(c.Returns.R, &returns.Return{
			R:    re,
			Date: c.I.Candles.C[len(c.Returns.R)].Date,
		})
	}

	return c
}

func (c *CAPM) CalculateAbnormalReturns() *EventStudy {
	es := &EventStudy{
		C: c,
		AR: &returns.Returns{
			R: []*returns.Return{},
		},
	}

	for i, r := range c.Returns.R {
		ab := c.I.Returns.R[i].R - r.R

		es.AR.R = append(es.AR.R, &returns.Return{
			R:    ab,
			Date: r.Date,
		})
	}

	es.AR.CalculateCumulative()

	return es
}

func (e *EventStudy) ToCsv(path string, del *rune) error {
	if del == nil {
		d := defaultDel
		del = &d
	}

	if err := os.MkdirAll(filepath.Dir(path), 0770); err != nil && !os.IsExist(err) {
		return err
	}

	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0770)

	if err != nil {
		return err
	}

	defer f.Close()

	_, err = f.WriteString("Date,R (real),R (forecasted),Rm,AR,CAR\n")

	if err != nil {
		return err
	}

	for i, r := range e.C.I.Returns.R {
		if len(e.AR.R)  <= i {
			break
		}

		var l string
		l += r.Date.Format(layout)
		l += string(*del)
		l += fmt.Sprintf("%f", r.R)
		l += string(*del)
		l += fmt.Sprintf("%f", e.C.Returns.R[i].R)
		l += string(*del)
		l += fmt.Sprintf("%f", e.C.Rm.Returns.R[i].R)
		l += string(*del)
		l += fmt.Sprintf("%f", e.AR.R[i].R)
		l += string(*del)
		l += fmt.Sprintf("%f", e.AR.CR[i].Value)
		l += "\n"

		if _, err = f.WriteString(l); err != nil {
			return err
		}
	}

	return nil
}
