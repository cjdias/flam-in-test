package test

import (
	"github.com/cjdias/flam-in-go"
)

func newErrNilReference(
	arg string,
) error {
	return flam.NewErrorFrom(
		flam.ErrNilReference,
		arg)
}
