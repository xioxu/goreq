package goreq

var trueVal = true
var falseVal = false

var TrueVal = &trueVal
var FalseVal = &falseVal

func NewString(str string)  *string{
	return &str
}
