// Package brand provides Go Brand colors.
package brand

import (
	"fmt"
	"sort"

	"github.com/lucasb-eyer/go-colorful"
)

// Brand colors in hex.
const (
	GopherBlue = "#00ADD8"
	LightBlue  = "#5DC9E2"
	Aqua       = "#00A29C"
	Black      = "#000000"
	Fuchsia    = "#CE3262"
	Yellow     = "#FDDD00"
	Turquoise  = "#00758D"
	Slate      = "#555759"
	Purple     = "#402B56"
	CoolGray   = "#DBD9D6"
)

// Colors is a map of named brand colors.
var Colors = map[string]string{
	"gopher-blue": GopherBlue,
	"light-blue":  LightBlue,
	"aqua":        Aqua,
	"black":       Black,
	"fuchsia":     Fuchsia,
	"yellow":      Yellow,
	"turquoise":   Turquoise,
	"slate":       Slate,
	"purple":      Purple,
	"cool-gray":   CoolGray,
}

// ColorNames is the list of color names.
var ColorNames []string

func init() {
	for name := range Colors {
		ColorNames = append(ColorNames, name)
	}
	sort.Strings(ColorNames)
}

// Color looks up a named color.
func Color(name string) (string, error) {
	hex, ok := Colors[name]
	if !ok {
		return "", fmt.Errorf("unknown color %q", name)
	}
	return hex, nil
}

// Lighten returns the named color lightened by the given proportion (in range 0-1).
func Lighten(name string, p float64) (string, error) {
	hx, err := Color(name)
	if err != nil {
		return "", err
	}

	c := hex(hx)
	h, s, l := c.Hsl()
	l = (1-p)*l + p

	return colorful.Hsl(h, s, l).Hex(), nil
}

func hex(h string) colorful.Color {
	c, err := colorful.Hex(h)
	if err != nil {
		panic(err)
	}
	return c
}
