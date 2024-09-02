package util

import (
	"errors"
	"fmt"
)

func Must(err error) {
	if err != nil {
		panic(err)
	}
}

func Must1[T any](val T, err error) T {
	if err != nil {
		panic(err)
	}
	return val
}

func Recover(err *error) {
	p := recover()
	if perr, ok := p.(error); ok {
		*err = perr
	} else if pstr, ok := p.(fmt.Stringer); ok {
		*err = errors.New(pstr.String())
	} else if pstr, ok := p.(string); ok {
		*err = errors.New(pstr)
	} else if p != nil {
		*err = errors.New(fmt.Sprint(p))
	}
}
