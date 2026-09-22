// Package swe is the small cgo boundary around the vendored Swiss Ephemeris.
package swe

/*
#cgo CFLAGS: -O2
#cgo LDFLAGS: -lm
#include <stdlib.h>
#include "swephexp.h"
*/
import "C"

import (
	"fmt"
	"unsafe"
)

const (
	Sun          = int(C.SE_SUN)
	Moon         = int(C.SE_MOON)
	Mercury      = int(C.SE_MERCURY)
	Venus        = int(C.SE_VENUS)
	Mars         = int(C.SE_MARS)
	Jupiter      = int(C.SE_JUPITER)
	Saturn       = int(C.SE_SATURN)
	TrueNode     = int(C.SE_TRUE_NODE)
	Gregorian    = int(C.SE_GREG_CAL)
	FlagSwiss    = int(C.SEFLG_SWIEPH)
	FlagSpeed    = int(C.SEFLG_SPEED)
	FlagSidereal = int(C.SEFLG_SIDEREAL)
	Rise         = int(C.SE_CALC_RISE)
	Set          = int(C.SE_CALC_SET)
)

func SetEphemerisPath(path string) {
	p := C.CString(path)
	defer C.free(unsafe.Pointer(p))
	C.swe_set_ephe_path(p)
}

func SetLahiri() { C.swe_set_sid_mode(C.SE_SIDM_LAHIRI, 0, 0) }

func Close() { C.swe_close() }

// Ayanamsa returns the configured (Lahiri) ayanamsa in degrees at a UT Julian day.
func Ayanamsa(jdUT float64) float64 { return float64(C.swe_get_ayanamsa_ut(C.double(jdUT))) }

func JulianDay(year, month, day int, hourUTC float64) float64 {
	return float64(C.swe_julday(C.int(year), C.int(month), C.int(day), C.double(hourUTC), C.int(Gregorian)))
}

func ReverseJulian(jd float64) (year, month, day int, hour float64) {
	var y, m, d C.int
	var h C.double
	C.swe_revjul(C.double(jd), C.int(Gregorian), &y, &m, &d, &h)
	return int(y), int(m), int(d), float64(h)
}

func Position(jd float64, body int) (longitude, speed float64, err error) {
	longitude, _, _, speed, err = Position3D(jd, body)
	return
}

func Position3D(jd float64, body int) (longitude, latitude, distance, speed float64, err error) {
	var xx [6]C.double
	var serr [256]C.char
	flags := C.int(FlagSwiss | FlagSpeed | FlagSidereal)
	ret := C.swe_calc_ut(C.double(jd), C.int(body), flags, &xx[0], &serr[0])
	if ret < 0 {
		return 0, 0, 0, 0, fmt.Errorf("swe_calc_ut: %s", C.GoString(&serr[0]))
	}
	if int(ret)&FlagSwiss == 0 {
		return 0, 0, 0, 0, fmt.Errorf("Swiss ephemeris data unavailable: %s", C.GoString(&serr[0]))
	}
	return float64(xx[0]), float64(xx[1]), float64(xx[2]), float64(xx[3]), nil
}

func Ascendant(jd, latitude, longitude float64) (float64, error) {
	var cusps [13]C.double
	var ascmc [10]C.double
	ret := C.swe_houses_ex(C.double(jd), C.int(FlagSidereal), C.double(latitude), C.double(longitude), C.int('W'), &cusps[0], &ascmc[0])
	if ret < 0 {
		return 0, fmt.Errorf("swe_houses_ex failed")
	}
	return float64(ascmc[0]), nil
}

func RiseSet(afterJD, latitude, longitude float64, body, event int) (float64, error) {
	// Panchang Moon events use the geometric disc centre, without refraction.
	if body == Moon {
		event |= int(C.SE_BIT_DISC_CENTER | C.SE_BIT_NO_REFRACTION)
	}
	geo := [3]C.double{C.double(longitude), C.double(latitude), 0}
	var result C.double
	var serr [256]C.char
	ret := C.swe_rise_trans(C.double(afterJD), C.int(body), nil,
		C.int(FlagSwiss), C.int(event), &geo[0], 0, 0, &result, &serr[0])
	if ret < 0 {
		return 0, fmt.Errorf("swe_rise_trans: %s", C.GoString(&serr[0]))
	}
	return float64(result), nil
}
