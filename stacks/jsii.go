package stacks

import (
	"github.com/aws/jsii-runtime-go"
)

func S(s string) *string { return jsii.String(s) }

func N(n int) *float64 { return jsii.Number(n) }

func B(b bool) *bool { return jsii.Bool(b) }
