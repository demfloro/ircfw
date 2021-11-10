package ircfw

type ircColor string

const (
	White      ircColor = "00"
	Black      ircColor = "01"
	Blue       ircColor = "02"
	Green      ircColor = "03"
	Red        ircColor = "04"
	Brown      ircColor = "05"
	Magenta    ircColor = "06"
	Orange     ircColor = "07"
	Yellow     ircColor = "08"
	LightGreen ircColor = "09"
	Cyan       ircColor = "10"
	LightCyan  ircColor = "11"
	LightBlue  ircColor = "12"
	Pink       ircColor = "13"
	Grey       ircColor = "14"
	LightGrey  ircColor = "15"
	Default    ircColor = "99"
	NoColor    ircColor = ""
)

const (
	ColorTag = "\x03"
)

var (
	colors = map[string]ircColor{
		"00": White,
		"01": Black,
		"02": Blue,
		"03": Green,
		"04": Red,
		"05": Brown,
		"06": Magenta,
		"07": Orange,
		"08": Yellow,
		"09": LightGreen,
		"10": Cyan,
		"11": LightCyan,
		"12": LightBlue,
		"13": Pink,
		"14": Grey,
		"15": LightGrey,
		"99": Default,
	}
)

func (c ircColor) String() string {
	return string(c)
}

func lookupColor(str string) ircColor {
	color, ok := colors[str]
	if !ok {
		return NoColor
	}
	return color
}
