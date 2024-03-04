package bash

import (
	"fmt"
	"math"
	"math/big"
	"sort"
	"strings"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
)

var scriptFunction = &function.Spec{
	Description: "Prepends a bash script with some variable declarations based on given values.",
	Params: []function.Parameter{
		{
			Name:        "script_src",
			Type:        cty.String,
			Description: "The source code of the script to generate, without the variable declarations inserted yet.",
		},
		{
			Name:        "variables",
			Type:        cty.DynamicPseudoType,
			Description: "An object whose attributes represent the variables to declare.",
		},
	},
	Type: function.StaticReturnType(cty.String),
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		source := args[0].AsString()
		varsVal := args[1]
		if !(varsVal.Type().IsObjectType() || varsVal.Type().IsMapType()) {
			return cty.UnknownVal(retType), function.NewArgErrorf(1, "must be an object whose attributes represent the bash variables to declare")
		}
		if !varsVal.IsWhollyKnown() {
			return cty.UnknownVal(retType), nil
		}
		varVals := varsVal.AsValueMap()
		names := make([]string, 0, len(varVals))
		for name := range varVals {
			names = append(names, name)
		}
		sort.Strings(names)

		for _, name := range names {
			val := varVals[name]
			if len(name) == 0 {
				return cty.UnknownVal(retType), function.NewArgErrorf(1, fmt.Sprintf("cannot use empty string as a bash variable name"))
			}
			if !validVariableName(name) {
				return cty.UnknownVal(retType), function.NewArgErrorf(1, fmt.Sprintf("cannot use %q as a bash variable name", name))
			}
			ty := val.Type()
			switch {
			case ty == cty.String:
				// all strings are allowed
			case ty == cty.Number:
				// must be something we can represent as an int64
				bf := val.AsBigFloat()
				if _, acc := bf.Int64(); acc != big.Exact {
					return cty.UnknownVal(retType), function.NewArgErrorf(1, fmt.Sprintf("invalid value for %q: must be a whole number between %d and %d", name, math.MinInt64, math.MaxInt64))
				}
			case cty.List(cty.String).Equals(ty) || cty.Map(cty.String).Equals(ty):
				for it := val.ElementIterator(); it.Next(); {
					_, v := it.Element()
					if v.IsNull() {
						return cty.UnknownVal(retType), function.NewArgErrorf(1, fmt.Sprintf("invalid value for %q: elements must not be null", name))
					}
				}
			default:
				// We can't support any other types
				return cty.UnknownVal(retType), function.NewArgErrorf(1, fmt.Sprintf("invalid value for %q: Bash supports only strings, whole numbers, lists of strings, and maps of strings", name))
			}
		}

		varDecls := variablesToBashDecls(varVals)
		var result string
		if strings.HasPrefix(source, "#!") {
			// If the source seems to start with an interpreter line then we'll
			// keep it at the start and insert the variables after it.
			newline := strings.Index(source, "\n")
			if newline < 0 {
				result = source + "\n" + varDecls
			} else {
				before, after := source[:newline+1], source[newline+1:]
				result = before + varDecls + after
			}
		} else {
			result = varDecls + source
		}

		return cty.StringVal(result), nil
	},
	RefineResult: func(rb *cty.RefinementBuilder) *cty.RefinementBuilder {
		// This function never produces a null result
		return rb.NotNull()
	},
}
