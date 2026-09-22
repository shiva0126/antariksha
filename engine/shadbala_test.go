package engine

import (
	"math"
	"strings"
	"testing"
)

func TestShadbalaReferenceChart(t *testing.T) {
	e := New("../ephe")
	c, err := e.BirthChart(ChartInput{Date: "1996-05-14", Time: "10:15", Lat: 12.97, Lon: 77.59, TZ: "Asia/Kolkata"})
	if err != nil {
		t.Fatal(err)
	}
	sb, err := e.Shadbala(c)
	if err != nil {
		t.Fatal(err)
	}
	if len(sb.Rows) != 7 {
		t.Fatalf("rows %d", len(sb.Rows))
	}
	rows := map[string]ShadbalaRow{}
	ranks := map[int]bool{}
	for _, r := range sb.Rows {
		rows[r.Graha] = r
		ranks[r.Rank] = true
		sum := r.Sthana + r.Dig + r.Kala + r.Chesta + r.Naisargika + r.Drik
		if math.Abs(sum-r.Total) > 1e-9 || math.Abs(r.Rupas-r.Total/60) > 1e-9 {
			t.Fatalf("%s totals inconsistent %+v", r.Graha, r)
		}
		if r.Dig < 0 || r.Dig > 60 || r.Details["uchcha"] < 0 || r.Details["uchcha"] > 60 || r.Chesta < 0 || r.Chesta > 60 {
			t.Fatalf("%s component out of range %+v", r.Graha, r)
		}
		if r.Rupas < 2 || r.Rupas > 14 {
			t.Fatalf("%s implausible %.2f rupas", r.Graha, r.Rupas)
		}
	}
	if len(ranks) != 7 {
		t.Fatalf("ranks not unique: %v", ranks)
	}
	// Sun at 29.86° Mesha is 160.1° from its fall point (190°).
	if u := rows["sun"].Details["uchcha"]; math.Abs(u-160.14/3) > 0.1 {
		t.Fatalf("sun uchcha %.2f", u)
	}
	// Mars in Mesha (moolatrikona 0–12°? no: 14.7°, so own sign) in the 10th: dig bala near maximum.
	if rows["mars"].Dig < 55 {
		t.Fatalf("mars in the 10th should have high dig bala: %.1f", rows["mars"].Dig)
	}
	// Tuesday, ~4.4 h after sunrise: weekday lord Mars, hora lord Moon.
	if rows["mars"].Details["abda_masa_vara_hora"] < 45 || rows["moon"].Details["abda_masa_vara_hora"] < 60 {
		t.Fatalf("vara/hora: mars %.1f moon %.1f; notes %v", rows["mars"].Details["abda_masa_vara_hora"], rows["moon"].Details["abda_masa_vara_hora"], sb.Notes)
	}
	if !strings.Contains(strings.Join(sb.Notes, " "), "Weekday lord Mars, hora lord Moon") {
		t.Fatalf("notes %v", sb.Notes)
	}
	// Mercury and Jupiter are retrograde at birth. Jupiter stationed on about
	// 4 May 1996, so ten days later it is anuvakra (just turned), not full vakra.
	for _, id := range []string{"mercury", "jupiter"} {
		if rows[id].Chesta < 30 || !strings.Contains(rows[id].ChestaMotion, "vakra") {
			t.Fatalf("retrograde %s chesta %.1f (%s)", id, rows[id].Chesta, rows[id].ChestaMotion)
		}
	}
}

func TestCompoundRelationshipTable(t *testing.T) {
	gs := map[string]Graha{"sun": {Longitude: 0}, "moon": {Longitude: 30}, "venus": {Longitude: 180}, "mercury": {Longitude: 180}}
	// Sun→Moon: natural friend, Moon 2nd from Sun (temporary friend) → adhi mitra.
	if v := compound("sun", "moon", gs); v != 22.5 {
		t.Errorf("adhi mitra %v", v)
	}
	// Sun→Venus: natural enemy, Venus 7th (temporary enemy) → adhi shatru.
	if v := compound("sun", "venus", gs); v != 1.875 {
		t.Errorf("adhi shatru %v", v)
	}
	// Sun→Mercury: natural neutral, 7th (temporary enemy) → shatru.
	if v := compound("sun", "mercury", gs); v != 3.75 {
		t.Errorf("shatru %v", v)
	}
	// Moon→Sun: natural friend, Sun 12th from Moon (temporary friend) → adhi mitra.
	if v := compound("moon", "sun", gs); v != 22.5 {
		t.Errorf("moon-sun %v", v)
	}
}

func TestDrishtiSpecialAspects(t *testing.T) {
	if v := drishti("venus", 0, 180); v != 60 {
		t.Errorf("full 7th aspect %v", v)
	}
	if v := drishti("saturn", 0, 60); v < 59 {
		t.Errorf("saturn 3rd aspect %v", v)
	}
	if v := drishti("jupiter", 0, 120); v < 59 {
		t.Errorf("jupiter 5th aspect %v", v)
	}
	if v := drishti("mars", 0, 90); v < 59 {
		t.Errorf("mars 4th aspect %v", v)
	}
	if v := drishti("venus", 0, 15); v != 0 {
		t.Errorf("no aspect within 30° %v", v)
	}
}
