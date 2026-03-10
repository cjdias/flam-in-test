package test

import (
	"github.com/cjdias/flam-in-go"
)

type configSource struct {
	priority int
	bag      flam.Bag
}

func (s configSource) Close() error {
	return nil
}

func (s configSource) GetPriority() int {
	return s.priority
}

func (s configSource) SetPriority(priority int) {
	s.priority = priority
}

func (s configSource) Has(
	path string,
) bool {
	return s.bag.Has(path)
}

func (s configSource) Get(
	path string,
	def ...any,
) any {
	return s.bag.Get(path, def...)
}
