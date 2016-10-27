package handler

import (
	"testing"
)

func TestCheckAmount(t *testing.T) {

	var ret uint = 0

	var rinput = []float64{0.1, 0, 1, 1.1, 1.10, 1.2, 2.1, 2.11, 4.11, 4.99, 5.00, 5.01, 99999999.99,
		0.111, 1.100, 1.111, 1.555, 4.111, 4.999, 5.001}
	var winput = []float64{-1.1, 100000000.00, 100000000.01}

	for _, v := range rinput {
		t.Log(v)
		ret = CheckAmount(v)
		if ret != 0 {
			t.Errorf("rinput:%v, err code:%v\n", v, ret)
		}
	}

	for _, v := range winput {
		ret = CheckAmount(v)
		if ret == 0 {
			t.Errorf("winput:%v, err code:%v\n", v, ret)
		}
	}
}
