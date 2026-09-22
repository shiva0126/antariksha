package places

import "testing"

func TestSearch(t *testing.T) {
	if Count() < 30000 {
		t.Fatalf("index has %d places", Count())
	}
	cases := map[string]string{"bengaluru": "Bengaluru", "bangalore": "Bengaluru", "bombay": "Mumbai", "udupi": "Udupi", "varanasi": "Varanasi", "new york": "New York City", "pondich": "Puducherry"}
	for q, want := range cases {
		r := Search(q, 5)
		if len(r) == 0 || r[0].Name != want {
			t.Errorf("%q → %+v, want %s", q, r, want)
		}
	}
	r := Search("aurangabad, bihar", 5)
	if len(r) == 0 || r[0].Region != "Bihar" {
		t.Errorf("region filter: %+v", r)
	}
	if r := Search("b", 5); len(r) != 0 {
		t.Error("single character query should return nothing")
	}
	if b := Search("Bengaluru", 1)[0]; b.TZ != "Asia/Kolkata" || b.Lat < 12.9 || b.Lat > 13.1 {
		t.Errorf("bengaluru %+v", b)
	}
}
