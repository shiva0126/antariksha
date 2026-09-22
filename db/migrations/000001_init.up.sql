CREATE TABLE locations (
 id BIGSERIAL PRIMARY KEY, name TEXT NOT NULL, country TEXT NOT NULL,
 latitude DOUBLE PRECISION NOT NULL CHECK (latitude BETWEEN -90 AND 90),
 longitude DOUBLE PRECISION NOT NULL CHECK (longitude BETWEEN -180 AND 180),
 timezone TEXT NOT NULL, UNIQUE(name,country)
);
CREATE INDEX idx_locations_latlon ON locations(latitude,longitude);

CREATE TABLE panchang_cache (
 id BIGSERIAL PRIMARY KEY, civil_date DATE NOT NULL,
 lat_key NUMERIC(6,2) NOT NULL, lon_key NUMERIC(6,2) NOT NULL,
 ayanamsa TEXT NOT NULL DEFAULT 'lahiri', payload JSONB NOT NULL,
 computed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
 UNIQUE(civil_date,lat_key,lon_key,ayanamsa)
);
CREATE INDEX idx_cache_lookup ON panchang_cache(civil_date,lat_key,lon_key);

CREATE TABLE festival_rules (
 id BIGSERIAL PRIMARY KEY, slug TEXT UNIQUE NOT NULL, display_name TEXT NOT NULL,
 rule_type TEXT NOT NULL, tithi INT CHECK(tithi BETWEEN 1 AND 15),
 paksha TEXT CHECK(paksha IN ('shukla','krishna')), lunar_month TEXT,
 region TEXT DEFAULT 'all', month_system TEXT DEFAULT 'amanta' CHECK(month_system IN ('amanta','purnimanta')), notes TEXT
);
CREATE TABLE festival_occurrences (
 id BIGSERIAL PRIMARY KEY, rule_id BIGINT NOT NULL REFERENCES festival_rules(id) ON DELETE CASCADE,
 occurs_on DATE NOT NULL, year INT NOT NULL, region TEXT NOT NULL DEFAULT 'all',
 UNIQUE(rule_id,occurs_on,region)
);
CREATE INDEX idx_festivals_date ON festival_occurrences(occurs_on);

CREATE TABLE users (id BIGSERIAL PRIMARY KEY,email TEXT UNIQUE,default_location_id BIGINT REFERENCES locations(id),created_at TIMESTAMPTZ DEFAULT now());
CREATE TABLE reminders (id BIGSERIAL PRIMARY KEY,user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,festival_slug TEXT NOT NULL,lead_hours INT DEFAULT 24,channel TEXT DEFAULT 'push',active BOOLEAN DEFAULT true);

