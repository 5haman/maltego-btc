package main

import (
	"os"
)

func main() {
  lt := ParseLocalArguments(os.Args)
	input := lt.Value

	InitCache()
	tr := GetTransform(input)
	TransformOut(&tr)
}
