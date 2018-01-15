# Minimal support of Kryo serialization for Golang

## Installation

Install:

```shell
go get -u github.com/idaunis/kryo
```

Import:

```go
import "github.com/idaunis/kryo"
```

## Quickstart
```go
package main

import (
	"fmt"
	"io/ioutil"
	"log"

	"github.com/idaunis/kryo"
)

func main() {
	input, err := ioutil.ReadFile("data.bin")
	if err != nil {
		log.Fatal(err)
	}

	k := kryo.New(input)

	sampleInt := k.ReadInt()
  sampleString := k.ReadString()

	fmt.Printf("Deserialized contents: %d %s", sampleInt, sampleString)
}
```
