package net

import (
	"fmt"
	"testing"
)

func TestCheckNodeState(t *testing.T) {
	res := CheckNodeState("127.0.0.1", "80")
	fmt.Println(res)
}
