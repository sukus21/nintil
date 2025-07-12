package util

import "io"

// If error == io.EOF, translate to io.EffUnexpectedEOF
func TranslateEOF(err error) error {
	if err == io.EOF {
		return io.ErrUnexpectedEOF
	} else {
		return err
	}
}
