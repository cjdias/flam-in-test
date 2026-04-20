package test

type arrangeReg struct {
	id       string
	function any
}

type assertReg struct {
	id       string
	function any
}

type configReg struct {
	file     string
	priority int
}

type psReg struct {
	channel string
	data    any
}

type teardownReg struct {
	id       string
	function any
}
