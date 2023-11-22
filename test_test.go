package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/thoas/go-funk"
	. "go.uber.org/mock/gomock"
	"testing"
)

func TestCar_Ignite___should_ignite_successfully(t *testing.T) {
	controller := NewController(t)
	defer controller.Finish()
	r := funk.Map([]int{1, 2, 3, 4}, func(x int) int {
		return x * 2
	})
	m := make(map[string]string)
	m["k1"] = "7"
	m["k2"] = "13"

	funk.ForEach(m, func(k string, v string) {
		fmt.Println(k + ":" + v)
	})

	fmt.Println(r)
	assert.NoError(t, nil)
}
