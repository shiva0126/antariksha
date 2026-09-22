package engine

import (
	"math"
	"sync"
	"testing"
)

// Reference: upstream Swiss Ephemeris swetest (separate C executable):
// -b14.5.1996 -ut4:45 -p0123456t -sid1 -fPlbs -g, -head -house77.59,12.97,W
// Birth instant is 10:15 IST = 04:45 UTC. Never use example JSON as a fixture.
func TestChartAgainstSwissReferenceConcurrent(t *testing.T) {
	e := New("../ephe")
	want := map[string]float64{"sun": 29.8635591, "moon": 350.5216950, "mercury": 31.1869963, "venus": 63.7989341, "mars": 14.7410520, "jupiter": 263.7047476, "saturn": 340.2224565, "rahu": 172.9698223, "ketu": 352.9698223}
	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c, err := e.BirthChart(ChartInput{Date: "1996-05-14", Time: "10:15", Lat: 12.97, Lon: 77.59, TZ: "Asia/Kolkata"})
			if err != nil {
				t.Error(err)
				return
			}
			for _, g := range c.Grahas {
				if math.Abs(g.Longitude-want[g.ID]) > 0.000001 {
					t.Errorf("%s: %.9f want %.9f", g.ID, g.Longitude, want[g.ID])
				}
			}
			if math.Abs(c.Ascendant.Longitude-90.4806822) > 0.000001 {
				t.Errorf("ascendant %.9f", c.Ascendant.Longitude)
			}
			if c.Ascendant.Rashi != "Karka" || c.Houses.Cusps[0] != 90 || c.Houses.Cusps[9] != 0 {
				t.Errorf("whole-sign boundaries: %+v", c.Houses)
			}
		}()
	}
	wg.Wait()
}
