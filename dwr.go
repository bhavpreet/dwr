package main

import (
	"sort"
)

const RedisWeightsKey = "DWR_WEIGHTS"

type kw struct {
	Key    string `json:"key"`
	Weight int    `json:"weight"`
}

type kwA []kw

type weightsBundle struct {
	UW map[string]int `json:"uw"` // userWeights
	DW kwA            `json:"dw"` // deducedWeights
	TW int            `json:"tw"` // total weight
}

type weights map[string]*weightsBundle

// sort interface
func (_kw kwA) Len() int {
	return len(_kw)
}

func (_kw kwA) Less(i, j int) bool {
	return _kw[i].Weight < _kw[j].Weight
}

func (_kw kwA) Swap(i, j int) {
	_kw[i], _kw[j] = _kw[j], _kw[i]
}

func (x *weightsBundle) ComputeWeights() {

	kwa := make(kwA, 0) // kw(a) is an array

	for k := range x.UW {
		kwa = append(kwa, kw{k, x.UW[k]})
	}

	// Sorting time
	sort.Stable(kwa)

	var _gcd int

	for idx, kw := range kwa {
		if kw.Weight != 0 {
			_gcd = deduceGcd(kw.Weight, kwa[idx+1:])
			break
		}
	}

	// fmt.Println("gcd : ", _gcd)
	// fmt.Println(kwa)
	for key := range x.UW {
		x.DW = append(x.DW, kw{key, x.UW[key] / _gcd})
		x.TW += x.UW[key] / _gcd
	}
}
