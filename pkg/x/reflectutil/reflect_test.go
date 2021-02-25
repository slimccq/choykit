// Copyright © 2021 ichenq@outlook.com. All Rights Reserved.
//
// Any redistribution or reproduction of part or all of the contents in any form
// is prohibited.
//
// You may not, except with our express written permission, distribute or commercially
// exploit the content. Nor may you transmit it or store it in any other website or
// other form of electronic retrieval system.

package reflectutil

import (
	"fmt"
	"testing"
)

type AA struct {
	B int
	C bool
	D float64
	F string
}

func TestGetStructAllFieldValues(t *testing.T) {
	var a = &AA{123, false, 3.14, "ok"}
	result := GetStructAllFieldValues(a)
	fmt.Printf("%v\n", result)
}

func TestGetStructFieldValues(t *testing.T) {
	var a = &AA{123, false, 3.14, "ok"}
	result := GetStructFieldValues(a, "D")
	fmt.Printf("%v\n", result)
}

func TestGetStructFieldValuesBy(t *testing.T) {
	var a = &AA{123, false, 3.14, "ok"}
	result := GetStructFieldValuesBy(a, []string{"B", "C", "F"})
	fmt.Printf("%v\n", result)
}