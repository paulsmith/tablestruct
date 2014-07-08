package main

import (
	"flag"
	"log"
	"os"

	"github.com/paulsmith/tablestruct"
)

func main() {
	var (
		pkg = flag.String("pkg", "main", "package of generated code")
	)

	flag.Parse()

	mapper, err := tablestruct.NewMap(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}

	code := tablestruct.NewCode()
	code.Gen(mapper, *pkg, os.Stdout)
}
