package main

import (
	"fmt"
	"os"

	"foobar"

	"github.com/cespare/argf"
	"github.com/cespare/hutil/apachelog"
)

func main() {
	_ = apachelog.ApacheTimeFormat
	foobar.Say()

	for argf.Scan() {
		fmt.Println(argf.String())
	}
	if err := argf.Error(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
