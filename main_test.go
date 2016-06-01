package main

import "testing"

var emptyGoodness = roundGoodness{}

var idealPairing = pairingDetails{rematch: 0, groupDiff: 0, sideDiffs: [2]int{0, 0}, streaks: [2]int{1, 1}}
var groupCrossingPairing = pairingDetails{0, 1, [2]int{0, 0}, [2]int{1, 1}}
var rematchPairing = pairingDetails{1, 0, [2]int{0, 0}, [2]int{1, 1}}
var sideDiffsPairing = pairingDetails{0, 0, [2]int{1, 2}, [2]int{1, 1}}
var streaksPairing = pairingDetails{0, 0, [2]int{0, 0}, [2]int{1, 2}}
var messyPairing = pairingDetails{0, 2, [2]int{2, 1}, [2]int{1, 2}}

var addPairingTests = []struct {
	in  roundGoodness
	p   pairingDetails
	out roundGoodness
}{
	{
		emptyGoodness,
		idealPairing,
		roundGoodness{[]int{1}, []int{1}, []int{2}, []int{0, 2}},
	},
	{
		emptyGoodness,
		groupCrossingPairing,
		roundGoodness{[]int{1}, []int{0, 1}, []int{2}, []int{0, 2}},
	},
	{
		emptyGoodness,
		rematchPairing,
		roundGoodness{[]int{0, 1}, []int{1}, []int{2}, []int{0, 2}},
	},
	{
		emptyGoodness,
		sideDiffsPairing,
		roundGoodness{[]int{1}, []int{1}, []int{0, 1, 1}, []int{0, 2}},
	},
	{
		emptyGoodness,
		streaksPairing,
		roundGoodness{[]int{1}, []int{1}, []int{2}, []int{0, 1, 1}},
	},
	{
		emptyGoodness,
		messyPairing,
		roundGoodness{[]int{1}, []int{0, 0, 1}, []int{0, 1, 1}, []int{0, 1, 1}},
	},
}

func (g1 *roundGoodness) equals(g2 *roundGoodness) bool {
	if len(g1.rematches) != len(g2.rematches) ||
		len(g1.sideDiffs) != len(g2.sideDiffs) ||
		len(g1.streaks) != len(g2.streaks) ||
		len(g1.groupDiffs) != len(g2.groupDiffs) {
		return false
	}

	for i := 0; i < len(g1.rematches); i++ {
		if g1.rematches[i] != g2.rematches[i] {
			return false
		}
	}

	for i := 0; i < len(g1.sideDiffs); i++ {
		if g1.sideDiffs[i] != g2.sideDiffs[i] {
			return false
		}
	}

	for i := 0; i < len(g1.streaks); i++ {
		if g1.streaks[i] != g2.streaks[i] {
			return false
		}
	}

	for i := 0; i < len(g1.groupDiffs); i++ {
		if g1.groupDiffs[i] != g2.groupDiffs[i] {
			return false
		}
	}

	return true
}

func TestAddPairing(t *testing.T) {
	for _, data := range addPairingTests {
		g := data.in
		g.addPairing(data.p)
		if !g.equals(&data.out) {
			t.Error("For", data.in,
				"plus", data.p,
				"expected", data.out,
				"got", g,
			)
		}
	}
}
