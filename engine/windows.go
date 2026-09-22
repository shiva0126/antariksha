package engine

import "time"

var rahuPart = [7]int{7, 1, 6, 4, 5, 3, 2}
var yamaPart = [7]int{4, 3, 2, 1, 0, 6, 5}
var gulikaPart = [7]int{6, 5, 4, 3, 2, 1, 0}

func slice(start, end time.Time, part, count int) (time.Time, time.Time) {
	d := end.Sub(start) / time.Duration(count)
	return start.Add(time.Duration(part) * d), start.Add(time.Duration(part+1) * d)
}
func window(start, end time.Time, part int, z *time.Location) Window {
	a, b := slice(start, end, part, 8)
	return Window{clock(a, z), clock(b, z)}
}

func applyWindows(day *Day, rise, set time.Time, z *time.Location, wd time.Weekday) {
	i := int(wd)
	day.RahuKaal = window(rise, set, rahuPart[i], z)
	day.Yamaganda = window(rise, set, yamaPart[i], z)
	day.Gulika = window(rise, set, gulikaPart[i], z)
	noon := rise.Add(set.Sub(rise) / 2)
	muhurta := set.Sub(rise) / 15
	day.Abhijit = Window{clock(noon.Add(-muhurta/2), z), clock(noon.Add(muhurta/2), z)}
	nextRise := rise.Add(24 * time.Hour)
	day.Choghadiya = append(choghadiya(rise, set, wd, true, z), choghadiya(set, nextRise, wd, false, z)...)
}

var dayChog = [7][8]string{{"Udveg", "Chara", "Labh", "Amrit", "Kala", "Shubh", "Rog", "Udveg"}, {"Amrit", "Kala", "Shubh", "Rog", "Udveg", "Chara", "Labh", "Amrit"}, {"Rog", "Udveg", "Chara", "Labh", "Amrit", "Kala", "Shubh", "Rog"}, {"Labh", "Amrit", "Kala", "Shubh", "Rog", "Udveg", "Chara", "Labh"}, {"Shubh", "Rog", "Udveg", "Chara", "Labh", "Amrit", "Kala", "Shubh"}, {"Chara", "Labh", "Amrit", "Kala", "Shubh", "Rog", "Udveg", "Chara"}, {"Kala", "Shubh", "Rog", "Udveg", "Chara", "Labh", "Amrit", "Kala"}}
var nightChog = [7][8]string{{"Shubh", "Amrit", "Chara", "Rog", "Kala", "Labh", "Udveg", "Shubh"}, {"Chara", "Rog", "Kala", "Labh", "Udveg", "Shubh", "Amrit", "Chara"}, {"Kala", "Labh", "Udveg", "Shubh", "Amrit", "Chara", "Rog", "Kala"}, {"Udveg", "Shubh", "Amrit", "Chara", "Rog", "Kala", "Labh", "Udveg"}, {"Amrit", "Chara", "Rog", "Kala", "Labh", "Udveg", "Shubh", "Amrit"}, {"Rog", "Kala", "Labh", "Udveg", "Shubh", "Amrit", "Chara", "Rog"}, {"Labh", "Udveg", "Shubh", "Amrit", "Chara", "Rog", "Kala", "Labh"}}

func choghadiya(start, end time.Time, wd time.Weekday, day bool, z *time.Location) []Choghadiya {
	out := make([]Choghadiya, 8)
	for i := range out {
		a, b := slice(start, end, i, 8)
		name := nightChog[int(wd)][i]
		if day {
			name = dayChog[int(wd)][i]
		}
		out[i] = Choghadiya{name, clock(a, z), clock(b, z)}
	}
	return out
}
