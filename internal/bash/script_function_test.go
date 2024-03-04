package bash

import (
	"testing"

	"github.com/zclconf/go-cty/cty"
)

func TestScriptFunction(t *testing.T) {
	tests := map[string]struct {
		Source    cty.Value
		Variables cty.Value
		Want      cty.Value
		WantErr   string
	}{
		"empty object": {
			Source:    cty.StringVal(""),
			Variables: cty.EmptyObjectVal,
			Want:      cty.StringVal(""),
		},
		"empty map": {
			Source:    cty.StringVal(""),
			Variables: cty.MapValEmpty(cty.String),
			Want:      cty.StringVal(""),
		},

		"string variable": {
			Source: cty.StringVal(`echo "$name"`),
			Variables: cty.ObjectVal(map[string]cty.Value{
				"name": cty.StringVal("Alex"),
			}),
			Want: cty.StringVal(`declare -r name='Alex'
echo "$name"`),
		},
		"integer variable": {
			Source: cty.StringVal(`echo "$num"`),
			Variables: cty.ObjectVal(map[string]cty.Value{
				"num": cty.NumberIntVal(12),
			}),
			Want: cty.StringVal(`declare -ri num=12
echo "$num"`),
		},
		"array variable": {
			Source: cty.StringVal(``),
			Variables: cty.ObjectVal(map[string]cty.Value{
				"names": cty.ListVal([]cty.Value{
					cty.StringVal("Alex"),
					cty.StringVal("Bitty"),
				}),
			}),
			Want: cty.StringVal(`declare -ra names=('Alex' 'Bitty')
`),
		},
		"empty array variable": {
			Source: cty.StringVal(``),
			Variables: cty.ObjectVal(map[string]cty.Value{
				"names": cty.ListValEmpty(cty.String),
			}),
			Want: cty.StringVal(`declare -ra names=()
`),
		},
		"associative array variable": {
			Source: cty.StringVal(``),
			Variables: cty.ObjectVal(map[string]cty.Value{
				"noises": cty.MapVal(map[string]cty.Value{
					"beep":  cty.StringVal("boop"),
					"bleep": cty.StringVal("bloop"),
				}),
			}),
			Want: cty.StringVal(`declare -rA noises=(['beep']='boop' ['bleep']='bloop')
`),
		},
		"empty associative array variable": {
			Source: cty.StringVal(``),
			Variables: cty.ObjectVal(map[string]cty.Value{
				"noises": cty.MapValEmpty(cty.String),
			}),
			Want: cty.StringVal(`declare -rA noises=()
`),
		},
		"many variables with interpreter line": {
			Source: cty.StringVal("#!/bin/bash\necho \"$name\"\n"),
			Variables: cty.ObjectVal(map[string]cty.Value{
				"name":  cty.StringVal("Alex"),
				"names": cty.ListValEmpty(cty.String),
				"noises": cty.MapVal(map[string]cty.Value{
					"beep":  cty.StringVal("boop"),
					"bleep": cty.StringVal("bloop"),
				}),
				"num": cty.NumberIntVal(12),
			}),
			Want: cty.StringVal(`#!/bin/bash
declare -r name='Alex'
declare -ra names=()
declare -rA noises=(['beep']='boop' ['bleep']='bloop')
declare -ri num=12
echo "$name"
`),
		},

		"invalid variable: tuple": {
			Source: cty.StringVal(""),
			Variables: cty.ObjectVal(map[string]cty.Value{
				"invalid": cty.EmptyTupleVal,
			}),
			WantErr: `invalid value for "invalid": Bash supports only strings, whole numbers, lists of strings, and maps of strings`,
		},
		"invalid variable: object": {
			Source: cty.StringVal(""),
			Variables: cty.ObjectVal(map[string]cty.Value{
				"invalid": cty.EmptyObjectVal,
			}),
			WantErr: `invalid value for "invalid": Bash supports only strings, whole numbers, lists of strings, and maps of strings`,
		},
		"invalid variable: bool": {
			Source: cty.StringVal(""),
			Variables: cty.ObjectVal(map[string]cty.Value{
				"invalid": cty.True,
			}),
			WantErr: `invalid value for "invalid": Bash supports only strings, whole numbers, lists of strings, and maps of strings`,
		},
		"invalid variable: list of number": {
			Source: cty.StringVal(""),
			Variables: cty.ObjectVal(map[string]cty.Value{
				"invalid": cty.ListValEmpty(cty.Number),
			}),
			WantErr: `invalid value for "invalid": Bash supports only strings, whole numbers, lists of strings, and maps of strings`,
		},
		"invalid variable: map of number": {
			Source: cty.StringVal(""),
			Variables: cty.ObjectVal(map[string]cty.Value{
				"invalid": cty.MapValEmpty(cty.Number),
			}),
			WantErr: `invalid value for "invalid": Bash supports only strings, whole numbers, lists of strings, and maps of strings`,
		},
		"invalid variable: set of string": {
			Source: cty.StringVal(""),
			Variables: cty.ObjectVal(map[string]cty.Value{
				"invalid": cty.SetValEmpty(cty.String),
			}),
			WantErr: `invalid value for "invalid": Bash supports only strings, whole numbers, lists of strings, and maps of strings`,
		},

		"invalid variables: string": {
			Source:    cty.StringVal(""),
			Variables: cty.StringVal("nope"),
			WantErr:   `must be an object whose attributes represent the bash variables to declare`,
		},
		"invalid variables: number": {
			Source:    cty.StringVal(""),
			Variables: cty.Zero,
			WantErr:   `must be an object whose attributes represent the bash variables to declare`,
		},
		"invalid variables: bool": {
			Source:    cty.StringVal(""),
			Variables: cty.True,
			WantErr:   `must be an object whose attributes represent the bash variables to declare`,
		},
		"invalid variables: list": {
			Source:    cty.StringVal(""),
			Variables: cty.ListValEmpty(cty.String),
			WantErr:   `must be an object whose attributes represent the bash variables to declare`,
		},
	}

	p := NewProvider()
	f := p.CallStub("script")
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := f(test.Source, test.Variables)

			if test.WantErr != "" {
				if err == nil {
					t.Fatalf("unexpected success\nwant error: %s", test.WantErr)
				}
				gotErr := err.Error()
				if gotErr != test.WantErr {
					t.Errorf("wrong error\ngot:  %s\nwant: %s", gotErr, test.WantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error\ngot error: %s\nwant: %#v", err, test.Want)
			}
			if !test.Want.RawEquals(got) {
				t.Fatalf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}
