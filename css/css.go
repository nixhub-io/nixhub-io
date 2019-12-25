package css

import (
	"html/template"
	"io"
	"strconv"
	"strings"
)

type Rules []Rule

func (rs *Rules) AddRule(r Rule) {
	*rs = append(*rs, r)
}

func (rs *Rules) SetRule(r Rule) {
	for i, rule := range *rs {
		if rule.Selector == r.Selector {
			(*rs)[i] = r
			return
		}
	}

	rs.AddRule(r)
}

func (rs *Rules) SetProperty(selector, property, value string) {
	for _, r := range *rs {
		if r.Selector == selector {
			r.SetProperty(property, value)
			return
		}
	}

	rs.AddRule(Rule{
		Selector:   selector,
		Properties: [][2]string{{property, value}},
	})
}

func (rs *Rules) CSS() string {
	var s strings.Builder
	for _, r := range *rs {
		s.WriteString(r.CSS())
	}
	return s.String()
}

func (rs *Rules) HTML() template.HTML {
	return WrapHTML(rs.CSS())
}

type Rule struct {
	Selector   string
	Properties [][2]string
}

func Content(selector, content string) Rule {
	return Single(selector, "content", Escape(content))
}

func Single(selector, key, value string) Rule {
	return Rule{
		Selector: selector,
		Properties: [][2]string{
			{key, value},
		},
	}
}

func (r *Rule) AddProperty(property, value string) {
	r.Properties = append(r.Properties, [2]string{property, value})
}

func (r *Rule) SetProperty(property, value string) {
	for i, prop := range r.Properties {
		if prop[0] == property {
			prop[1] = value

			// ???
			r.Properties[i] = prop
			return
		}
	}

	r.AddProperty(property, value)
}

func (r Rule) HTML() template.HTML {
	return WrapHTML(r.CSS())
}

func (r Rule) CSS() string {
	var s = r.Selector + " {"
	for _, prop := range r.Properties {
		s += prop[0] + ":" + prop[1] + ";"
	}
	return s + "}\n"
}

func WrapHTML(css string) template.HTML {
	return template.HTML(
		`<style type="text/css">` + css + `</style>`)
}

func WrapHTMLTo(w io.Writer, css string) error {
	_, err := w.Write([]byte(WrapHTML(css)))
	return err
}

// Escape returns the string escaped and quoted in double quotes.
func Escape(s string) string {
	var (
		runes   = []rune(s)
		escaped = make([]rune, 0, len(runes))
		buffer  = make([]rune, 0, 6)
	)

	for i := 0; i < len(runes); i++ {
		if badRune(runes[i]) {
			escaped = append(escaped, '\\')
			escaped = append(escaped, escape(runes[i], buffer)...)
			escaped = append(escaped, ' ')
		} else {
			escaped = append(escaped, runes[i])
		}
	}

	return `"` + string(escaped) + `"`
}

func badRune(r rune) bool {
	return r < 0x20 || r > 0x7E ||
		r == '"' || r == '\'' || r == '\\'
}

func escape(r rune, buffer []rune) []rune {
	buffer = buffer[:0]

	if r < 0xFF {
		return padRunes(
			[]rune(strconv.FormatUint(uint64(r), 16)),
			buffer, '0', 2,
		)
	}

	if r > 0xFFFFFF {
		return nil
	}

	return padRunes(
		[]rune(strconv.FormatUint(uint64(r), 16)),
		buffer, '0', 6,
	)
}

func padRunes(s, buffer []rune, char rune, max int) []rune {
	if len(s) >= max {
		return s
	}

	for i := len(s); i < max; i++ {
		buffer = append(buffer, char)
	}

	return append(buffer, s...)
}
