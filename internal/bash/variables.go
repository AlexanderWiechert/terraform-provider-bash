package bash

import (
	"fmt"
	"sort"
	"strings"

	"github.com/zclconf/go-cty/cty"
)

// variablesToBashDecls tries to produce a bash script fragment containing
// declarations for each of the variables described in vars.
//
// Only a subset of possible Terraform values can be translated to bash
// variables because of differences in type system, but this function assumes
// that the variable names and values were already checked during configuration
// decoding and so will just return something invalid if given an unsupported
// value to deal with.
func variablesToBashDecls(vars map[string]cty.Value) string {
	if len(vars) == 0 {
		return ""
	}

	var buf strings.Builder
	names := make([]string, 0, len(vars))
	for name := range vars {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		val := vars[name]
		ty := val.Type()
		switch {
		case ty == cty.String:
			s := val.AsString()
			buf.WriteString("declare -r ")
			buf.WriteString(name)
			buf.WriteString("=")
			buf.WriteString(bashQuoteString(s))
			buf.WriteString("\n")
		case ty == cty.Number:
			f := val.AsBigFloat()
			// NOTE: Bash only actually supports integers, so here we're
			// assuming that the caller already rejected fractional values.
			buf.WriteString("declare -ri ")
			buf.WriteString(name)
			buf.WriteString("=")
			buf.WriteString(f.Text('f', -1))
			buf.WriteString("\n")
		case cty.List(cty.String).Equals(ty):
			l := val.AsValueSlice()
			buf.WriteString("declare -ra ")
			buf.WriteString(name)
			buf.WriteString("=(")
			for i, ev := range l {
				es := ev.AsString()
				if i != 0 {
					buf.WriteString(" ")
				}
				buf.WriteString(bashQuoteString(es))
			}
			buf.WriteString(")\n")
		case cty.Map(cty.String).Equals(ty):
			m := val.AsValueMap()
			buf.WriteString("declare -rA ")
			buf.WriteString(name)
			buf.WriteString("=(")
			names := make([]string, 0, len(m))
			for name := range m {
				names = append(names, name)
			}
			sort.Strings(names)
			i := 0
			for _, ek := range names {
				ev := m[ek]
				es := ev.AsString()
				if i != 0 {
					buf.WriteString(" ")
				}
				buf.WriteString("[")
				buf.WriteString(bashQuoteString(ek))
				buf.WriteString("]=")
				buf.WriteString(bashQuoteString(es))
				i++
			}
			buf.WriteString(")\n")
		default:
			// Shouldn't get here if config decoding validation is working
			fmt.Fprintf(&buf, "# ERROR: Don't know how to serialize %q for bash\n", name)
		}
	}
	return buf.String()
}

func bashQuoteString(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

func validVariableName(s string) bool {
	if len(s) == 0 {
		return false
	}
	for i, c := range s {
		if i == 0 {
			if !validVariableNameInitialCharacter(c) {
				return false
			}
		} else {
			if !validVariableNameSubsequentCharacter(c) {
				return false
			}
		}
	}
	return true
}

func validVariableNameInitialCharacter(c rune) bool {
	switch {
	case c == '_':
		return true
	case c >= 'A' && c <= 'Z':
		return true
	case c >= 'a' && c <= 'z':
		return true
	default:
		return false
	}
}

func validVariableNameSubsequentCharacter(c rune) bool {
	switch {
	case validVariableNameInitialCharacter(c):
		return true
	case c >= '0' && c <= '9':
		return true
	default:
		return false
	}
}
