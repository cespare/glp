package main

import (
	"foobar"

	"github.com/cespare/argf"
	"github.com/cespare/hutil/apachelog"
)

func main() {
	_ = apachelog.ApacheTimeFormat
	_ = argf.Scan
	foobar.Say()
}
