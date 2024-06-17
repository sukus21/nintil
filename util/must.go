package util

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

func Recover(fn func(any)) {
	err := recover()
	if err != nil {
		fn(err)
	}
}
