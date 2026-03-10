package test

import (
	"fmt"

	"github.com/cjdias/flam-in-go"
)

type Logger interface {
	ConfigStart()
	ConfigGetSource(file string)
	ConfigRegSource(file string, data flam.Bag)
	ConfigEnd()

	BootStart()
	BootError(err error)
	BootEnd()

	ArrangeStart()
	ArrangeRun(id string)
	ArrangeError(id string, err error)
	ArrangeEnd()

	AppRunStart()
	AppRunError(err error)
	AppRunEnd()

	ActStart()
	ActError(err error)
	ActEnd()

	AssertStart()
	AssertRun(id string)
	AssertError(id string, err error)
	AssertEnd()

	TeardownStart()
	TeardownRun(id string)
	TeardownError(id string, err error)
	TeardownEnd()
}

type logger struct{}

func NewLogger() Logger {
	return &logger{}
}

func (logger) ConfigStart() {
	fmt.Println("[Runner] starting config ...")
}

func (logger) ConfigGetSource(
	file string,
) {
	fmt.Println("[Runner] retrieving config source:", file, "...")
}

func (logger) ConfigRegSource(
	file string,
	data flam.Bag,
) {
	fmt.Println("[Runner] registering config source:", file, "->", data)
}

func (logger) ConfigEnd() {
	fmt.Println("[Runner] config ended")
}

func (logger) BootStart() {
	fmt.Println("[Runner] starting boot ...")
}

func (logger) BootError(
	err error,
) {
	fmt.Println("[Runner] error running boot:", err)
}

func (logger) BootEnd() {
	fmt.Println("[Runner] boot ended")
}

func (logger) ArrangeStart() {
	fmt.Println("[Runner] starting arrange ...")
}

func (logger) ArrangeRun(id string) {
	fmt.Println("[Runner] running arrange:", id, "...")
}

func (logger) ArrangeError(id string, err error) {
	fmt.Println("[Runner] error running arrange:", id, "->", err)
}

func (logger) ArrangeEnd() {
	fmt.Println("[Runner] arrange ended")
}

func (logger) AppRunStart() {
	fmt.Println("[Runner] starting application ...")
}

func (logger) AppRunError(
	err error,
) {
	fmt.Println("[Runner] error running application:", err)
}

func (logger) AppRunEnd() {
	fmt.Println("[Runner] application ended")
}

func (logger) ActStart() {
	fmt.Println("[Runner] starting act ...")
}

func (logger) ActError(
	err error,
) {
	fmt.Println("[Runner] error running act:", err)
}

func (logger) ActEnd() {
	fmt.Println("[Runner] act ended")
}

func (logger) AssertStart() {
	fmt.Println("[Runner] starting assertions ...")
}

func (logger) AssertRun(id string) {
	fmt.Println("[Runner] running assertion:", id, "...")
}

func (logger) AssertError(id string, err error) {
	fmt.Println("[Runner] error running assertion:", id, "->", err)
}

func (logger) AssertEnd() {
	fmt.Println("[Runner] assertions ended")
}

func (logger) TeardownStart() {
	fmt.Println("[Runner] starting teardowns ...")
}

func (logger) TeardownRun(id string) {
	fmt.Println("[Runner] running teardown:", id, "...")
}

func (logger) TeardownError(id string, err error) {
	fmt.Println("[Runner] error running teardown:", id, "->", err)
}

func (logger) TeardownEnd() {
	fmt.Println("[Runner] teardowns ended")
}
