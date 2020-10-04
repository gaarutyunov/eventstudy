package returns

import (
	"time"
)

type Return struct {
	R    float64
	Date time.Time
}

type Moving struct {
	Date  time.Time
	Value float64
}

type Returns struct {
	R []*Return
	// Moving Average
	MA []*Moving
	// Moving Average Squared
	MAS []*Moving
	// Moving Variance
	MV []*Moving
	// Moving Covariance
	MCV []*Moving
	// Moving Beta
	MB []*Moving
	// Cumulative Returns
	CR []*Moving
}

func (c *Returns) Len() int {
	return len(c.R)
}

func (c *Returns) Less(i, j int) bool {
	return c.R[i].Date.Before(c.R[j].Date)
}

func (c *Returns) Swap(i, j int) {
	c.R[i], c.R[j] = c.R[j], c.R[i]
}

func (c *Returns) Beta() float64 {
	return c.MCV[len(c.MCV) - 1].Value / c.MV[len(c.MV) - 1].Value
}

func (c *Returns) CalculateCumulative() *Returns {
	if c.CR == nil {
		c.CR = []*Moving{}
	}
	for i, r := range c.R {
		if i == 0 {
			c.CR = append(c.CR, &Moving{
				Date:  r.Date,
				Value: 0,
			})
			continue
		}

		c.CR = append(c.CR, &Moving{
			Date:  r.Date,
			Value: c.R[i-1].R + r.R,
		})
	}

	return c
}

func (c *Returns) CalculateMovingAverage(window int) *Returns {
	if c.MA == nil {
		c.MA = []*Moving{}
	}
	if c.MAS == nil {
		c.MAS = []*Moving{}
	}

	for i, candle := range c.R {
		if i == 0 {
			c.MA = append(c.MA, &Moving{
				Date:  candle.Date,
				Value: candle.R,
			})
			c.MAS = append(c.MAS, &Moving{
				Date:  candle.Date,
				Value: candle.R,
			})

			continue
		}

		if window > i+1 {
			c.AppendMovingAverage(i+1, i)
		} else {
			c.AppendMovingAverage(window, i)
		}
	}

	return c
}

func (c *Returns) CalculateMovingVariance(window int) *Returns {
	if c.MV == nil {
		c.MV = []*Moving{}
	}

	for i, candle := range c.R {
		if i == 0 {
			c.MV = append(c.MV, &Moving{
				Date:  candle.Date,
				Value: 0,
			})

			continue
		}

		if window > i+1 {
			c.AppendMovingVariance(i)
		} else {
			c.AppendMovingVariance(i)
		}
	}

	return c
}

func (c *Returns) CalculateMovingCovariance(window int, rm *Returns) *Returns {
	if c.MCV == nil {
		c.MCV = []*Moving{}
	}

	for i, candle := range c.R {
		if i == 0 {
			c.MCV = append(c.MCV, &Moving{
				Date:  candle.Date,
				Value: 0,
			})

			continue
		}

		if window > i+1 {
			c.AppendMovingCovariance(i+1, rm, i)
		} else {
			c.AppendMovingCovariance(window, rm, i)
		}
	}

	return c
}

func (c *Returns) AppendMovingCovariance(window int, rm *Returns, i int) *Returns {
	p := c.MCV[i-1].Value
	candle := c.R[i]
	// equity
	old := c.R[i+1-window].R
	newC := candle.R
	oldMa := c.MA[i+1-window].Value
	newMa := c.MA[i].Value
	// market portfolio
	oldRm := rm.R[i+1-window].R
	newCRm := rm.R[i].R
	oldMaRm := rm.MA[i+1-window].Value
	newMaRm := rm.MA[i].Value

	c.MCV = append(c.MCV, &Moving{
		Date:  candle.Date,
		Value: movingCovariance(p, old, oldMa, oldRm, oldMaRm, newC, newMa, newCRm, newMaRm, window),
	})

	return c
}

func (c *Returns) AppendMovingAverage(window int, i int) *Returns {
	p := c.MA[i-1].Value
	candle := c.R[i]
	old := c.R[i+1-window].R
	newC := candle.R
	c.MA = append(c.MA, &Moving{
		Date:  candle.Date,
		Value: movingAverage(p, old, newC, window),
	})
	c.MAS = append(c.MAS, &Moving{
		Date:  candle.Date,
		Value: movingAverage(p, old*old, newC*newC, window),
	})

	return c
}

func (c *Returns) AppendMovingVariance(i int) *Returns {
	candle := c.R[i]
	newMA := c.MA[i].Value
	newMAS := c.MAS[i].Value

	c.MV = append(c.MV, &Moving{
		Date:  candle.Date,
		Value: movingVariance(newMA, newMAS),
	})

	return c
}

func movingAverage(prev, old, new float64, n int) float64 {
	return prev + (new-old)/float64(n)
}

func movingVariance(newMA, newMAS float64) float64 {
	return newMAS - newMA*newMA
}

func movingCovariance(prev, oldX, oldMAX, oldY, oldMAY, newX, newMAX, newY, newMAY float64, n int) float64 {
	return (prev*(float64(n)-1) - (oldX-oldMAX)*(oldY-oldMAY) + (newX-newMAX)*(newY-newMAY)) / float64(n)
}
