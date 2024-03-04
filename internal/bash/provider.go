package bash

import (
	"github.com/apparentlymart/go-tf-func-provider/tffunc"
)

func NewProvider() *tffunc.Provider {
	p := tffunc.NewProvider()
	p.AddFunction("script", scriptFunction)
	return p
}
