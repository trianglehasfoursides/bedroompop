package main

import (
	"fmt"

	"github.com/tidwall/gjson"
	"github.com/trianglehasfoursides/mathrock/node/server"
)

func main() {
	msg := server.ParseMessage([]byte(`{"key":"ping","value":"hello"}`))
	fmt.Println(gjson.Get(string(msg.Value), "value").String())
}
