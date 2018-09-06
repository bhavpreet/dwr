package main

import "testing"

func TestComputeWeights(t *testing.T) {
	wb := new(weightsBundle)
	wb.UW = make(map[string]int)
	wb.DW = make(kwA, 0)
	wb.UW["k1"] = 9
	wb.UW["k2"] = 3

	wb.ComputeWeights()

	for _, v := range wb.DW {
		if v.key == "k1" && v.weight != 3 {
			t.Error("k!=3")
		}
	}
}
