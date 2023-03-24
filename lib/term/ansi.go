package term

import (
	"fmt"
	"regexp"
	"strings"
	"text/template"
	"unicode"
	"unicode/utf8"
)

const (
	ansiPat       = `\033\[(([0-9]+;?)*[a-zA-Z]?)`
	rgbfgcolor    = "\033[38;5;%vm"
	rgbbgcolor    = "\033[48;5;%vm"
	truefgcolor   = "\033[38;2;%v;%v;%vm"
	truebgcolor   = "\033[48;2;%v;%v;%vm"
	clearLastLine = "\033[G\033[1A\033[K"
)

var spinGlyphs = []rune("⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏")

var funcMap = template.FuncMap{
	"bright":                  ansiStyler("3", "9"),
	"Bright":                  ansiStyler("4", "10"),
	"bold":                    ansiStyler("1"),
	"faint":                   ansiStyler("2"),
	"italic":                  ansiStyler("3"),
	"underline":               ansiStyler("4"),
	"invert":                  ansiStyler("7"),
	"black":                   ansiStyler("30"),
	"red":                     ansiStyler("31"),
	"green":                   ansiStyler("32"),
	"yellow":                  ansiStyler("33"),
	"blue":                    ansiStyler("34"),
	"magenta":                 ansiStyler("35"),
	"cyan":                    ansiStyler("36"),
	"white":                   ansiStyler("37"),
	"Black":                   ansiStyler("40"),
	"Red":                     ansiStyler("41"),
	"Green":                   ansiStyler("42"),
	"Yellow":                  ansiStyler("43"),
	"Blue":                    ansiStyler("44"),
	"Magenta":                 ansiStyler("45"),
	"Cyan":                    ansiStyler("46"),
	"White":                   ansiStyler("47"),
	"spin":                    spin,
	"trimTrailingWhitespaces": trimRightSpace,
	"rpad":                    rpad,
}

var spinIndex int

func spin() string {
	if spinIndex++; spinIndex >= len(spinGlyphs) {
		spinIndex = 0
	}
	return string(spinGlyphs[spinIndex])
}

type ansiStr struct {
	str  string
	vals []string
}

func parseAnsiString(str string) ansiStr {
	start := regexp.MustCompile("^" + ansiPat)
	end := regexp.MustCompile(ansiPat + "$")
	points := []string{}
	if start.MatchString(str) {
		ansiCmd := start.FindStringSubmatch(str)
		str = end.ReplaceAllString(start.ReplaceAllString(str, ""), "")
		points = strings.Split(ansiCmd[1][:len(ansiCmd[1])-1], ";")
		if len(points) == 1 && points[0] == "" {
			points = []string{}
		}
	}
	return ansiStr{str: str, vals: points}
}

func (ansi *ansiStr) add(a string) {
	ansi.vals = append(ansi.vals, a)
}

func (ansi *ansiStr) replace(before, after string) {
	for i, val := range ansi.vals {
		if strings.HasPrefix(val, before) && len(val) > len(before) {
			ansi.vals[i] = strings.Replace(val, before, after, 1)
			break
		}
	}
}

func (ansi ansiStr) String() string {
	if len(ansi.vals) == 0 {
		return ansi.str
	}
	return fmt.Sprintf("\033[%vm%v\033[m", strings.Join(ansi.vals, ";"), ansi.str)
}

func ansiStyler(attrs ...string) func(interface{}) string {
	return func(v interface{}) string {
		ansistr := parseAnsiString(fmt.Sprintf("%v", v))
		if len(attrs) == 1 {
			ansistr.add(attrs[0])
		} else if len(attrs) == 2 {
			ansistr.replace(attrs[0], attrs[1])
		}
		return ansistr.String()
	}
}

func removeANSI(src []byte) []byte {
	regx := regexp.MustCompile(ansiPat)
	return regx.ReplaceAll(src, []byte(""))
}

func wrapANSI(src string, width int) string {
	str := []byte(src)
	var output, currentLine []byte
	for _, s := range str {
		currentLine = append(currentLine, s)
		runes := utf8.RuneCount(removeANSI(currentLine))
		if s == '\n' || runes >= width-1 {
			if s != '\n' {
				currentLine = append(currentLine, '\n')
			}
			output = append(output, currentLine...)
			currentLine = []byte{}
		}
	}
	return string(append(output, currentLine...))
}

func rpad(s string, padding int) string {
	formattedString := fmt.Sprintf("%%-%ds", padding)
	return fmt.Sprintf(formattedString, s)
}

func trimRightSpace(s string) string {
	return strings.TrimRightFunc(s, unicode.IsSpace)
}
