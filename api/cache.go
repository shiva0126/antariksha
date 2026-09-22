package api

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"time"

	"github.com/example/panchang/engine"
	"github.com/example/panchang/reading"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Cache interface {
	Get(context.Context, time.Time, float64, float64) (engine.Day, bool, error)
	Put(context.Context, time.Time, float64, float64, engine.Day) error
}
type NoCache struct{}

func (NoCache) Get(context.Context, time.Time, float64, float64) (engine.Day, bool, error) {
	return engine.Day{}, false, nil
}
func (NoCache) Put(context.Context, time.Time, float64, float64, engine.Day) error { return nil }

type PostgresCache struct{ Pool *pgxpool.Pool }
type Festival struct {
	Slug string `json:"slug"`
	Name string `json:"name"`
	Date string `json:"date"`
}
type CachedReading struct {
	Facts   engine.ChartFacts
	Reading reading.Reading
	Model   string
}

func key(v float64) float64 { return math.Round(v*100) / 100 }
func (c PostgresCache) Get(ctx context.Context, d time.Time, lat, lon float64) (engine.Day, bool, error) {
	var raw []byte
	err := c.Pool.QueryRow(ctx, `SELECT payload FROM panchang_cache WHERE civil_date=$1 AND lat_key=$2 AND lon_key=$3 AND ayanamsa='lahiri'`, d.Format("2006-01-02"), key(lat), key(lon)).Scan(&raw)
	if errors.Is(err, pgx.ErrNoRows) {
		return engine.Day{}, false, nil
	}
	if err != nil {
		return engine.Day{}, false, err
	}
	var day engine.Day
	if err = json.Unmarshal(raw, &day); err != nil {
		return day, false, err
	}
	if day.SchemaVersion != engine.SchemaVersion {
		return day, false, nil
	}
	return day, true, nil
}
func (c PostgresCache) Put(ctx context.Context, d time.Time, lat, lon float64, day engine.Day) error {
	raw, e := json.Marshal(day)
	if e != nil {
		return e
	}
	_, e = c.Pool.Exec(ctx, `INSERT INTO panchang_cache(civil_date,lat_key,lon_key,ayanamsa,payload) VALUES($1,$2,$3,'lahiri',$4) ON CONFLICT(civil_date,lat_key,lon_key,ayanamsa) DO UPDATE SET payload=EXCLUDED.payload,computed_at=now()`, d.Format("2006-01-02"), key(lat), key(lon), raw)
	return e
}
func (c PostgresCache) Festivals(ctx context.Context, year int, region string) ([]Festival, error) {
	rows, e := c.Pool.Query(ctx, `SELECT r.slug,r.display_name,o.occurs_on::text FROM festival_occurrences o JOIN festival_rules r ON r.id=o.rule_id WHERE o.year=$1 AND ($2='' OR o.region=$2 OR o.region='all') ORDER BY o.occurs_on,r.display_name`, year, region)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []Festival{}
	for rows.Next() {
		var f Festival
		if e = rows.Scan(&f.Slug, &f.Name, &f.Date); e != nil {
			return nil, e
		}
		out = append(out, f)
	}
	return out, rows.Err()
}

func (c PostgresCache) GetReading(ctx context.Context, hash string) (CachedReading, bool, error) {
	var factsRaw, readingRaw []byte
	var model string
	err := c.Pool.QueryRow(ctx, `SELECT facts,reading,model FROM reading_cache WHERE chart_hash=$1`, hash).Scan(&factsRaw, &readingRaw, &model)
	if errors.Is(err, pgx.ErrNoRows) {
		return CachedReading{}, false, nil
	}
	if err != nil {
		return CachedReading{}, false, err
	}
	var facts engine.ChartFacts
	var out reading.Reading
	if err = json.Unmarshal(factsRaw, &facts); err != nil {
		return CachedReading{}, false, err
	}
	if err = json.Unmarshal(readingRaw, &out); err != nil {
		return CachedReading{}, false, err
	}
	return CachedReading{facts, out, model}, true, nil
}
func (c PostgresCache) PutReading(ctx context.Context, hash string, x CachedReading) error {
	facts, _ := json.Marshal(x.Facts)
	out, _ := json.Marshal(x.Reading)
	_, err := c.Pool.Exec(ctx, `INSERT INTO reading_cache(chart_hash,facts,reading,model) VALUES($1,$2,$3,$4) ON CONFLICT(chart_hash) DO UPDATE SET facts=EXCLUDED.facts,reading=EXCLUDED.reading,model=EXCLUDED.model`, hash, facts, out, x.Model)
	return err
}
