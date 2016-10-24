package main

import "testing"

// Game tests

var c = Player{Name: "Alice", Corp: "EtF", Runner: "Noise"}
var r = Player{Name: "Bob", Corp: "PE", Runner: "Mac"}
var gameTests = []struct {
	corp           *Player
	runner         *Player
	winner         *Player
	timed          bool
	corpPrestige   int
	runnerPrestige int
	desc           string
}{
	{&c, &r, &c, false, 3, 0, "Corp win"},
	{&c, &r, &c, true, 2, 0, "Corp timed win"},
	{&c, &r, nil, false, 1, 1, "Tie"},
	{&c, &r, &r, true, 0, 2, "Runner timed win"},
	{&c, &r, &r, false, 0, 3, "Runner win"},
}

func TestMatches(t *testing.T) {
	for _, data := range gameTests {
		m := Match{Game{Pairing: Pairing{Corp: data.corp, Runner: data.runner}}}
		m.Game.RecordResult(data.winner, data.timed)
		if !m.IsDone() {
			t.Error("Match with result recorded returned false for IsDone()")
		}
		cp := m.Game.CorpPrestige()
		rp := m.Game.RunnerPrestige()
		if cp != data.corpPrestige || rp != data.runnerPrestige {
			t.Error("For", data.desc,
				"expected corp prestige", data.corpPrestige,
				"and runner prestige", data.runnerPrestige,
				"got corp prestige", cp,
				"and runner prestige", rp,
			)
		}
		mcp := m.GetPrestige(data.corp)
		if mcp != cp {
			t.Error("For", data.desc,
				"game prestige", cp,
				"and match prestige", mcp,
				"for corp player did not match",
			)
		}
		mrp := m.GetPrestige(data.runner)
		if mrp != rp {
			t.Error("For", data.desc,
				"game prestige", rp,
				"and match prestige", mrp,
				"for runner player did not match",
			)
		}
		invalidp := m.GetPrestige(&Player{})
		invalidop := m.GetOpponent(&Player{})
		if invalidp != 0 {
			t.Error("For", data.desc,
				"got prestige", invalidp,
				"and opponent", invalidop,
				"for non-participating player; expected 0 and nil",
			)
		}
		winner := m.GetWinner()
		if winner != data.winner {
			t.Error("For", data.desc,
				"got winner", winner,
				"expected", data.winner,
			)
		}
		corpop := m.GetOpponent(data.corp)
		if corpop != data.runner {
			t.Error("For", data.desc,
				"got corp's opponent", corpop,
				"expected", data.runner,
			)
		}
		runnerop := m.GetOpponent(data.runner)
		if runnerop != data.corp {
			t.Error("For", data.desc,
				"got runner's opponent", runnerop,
				"expected", data.corp,
			)
		}
	}
}

func TestUnfinishedGame(t *testing.T) {
	g := &Game{Pairing: Pairing{Corp: &c, Runner: &r}}
	cp := g.CorpPrestige()
	rp := g.RunnerPrestige()
	if cp != 0 || rp != 0 {
		t.Error("For unfinished game, expected corp prestige 0 and runner prestige 0, got corp prestige", cp, "and runner prestige", rp)
	}
}

// Round goodness tests

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

func copyGoodness(g *roundGoodness) roundGoodness {
	var c roundGoodness
	c.rematches = append([]int(nil), g.rematches...)
	c.sideDiffs = append([]int(nil), g.sideDiffs...)
	c.streaks = append([]int(nil), g.streaks...)
	c.groupDiffs = append([]int(nil), g.groupDiffs...)
	return c
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
		g := copyGoodness(&data.in)
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

// goodness for ideal round with 6 players paired
var idealRoundGoodness = roundGoodness{[]int{3}, []int{3}, []int{6}, []int{0, 6}}

// these are ideal except in one aspect, still with 6 players paired
var rematchGoodness = roundGoodness{[]int{2, 1}, []int{3}, []int{6}, []int{0, 6}}
var multiRematchGoodness = roundGoodness{[]int{1, 2}, []int{3}, []int{6}, []int{0, 6}}
var groupDiffGoodness = roundGoodness{[]int{3}, []int{2, 1}, []int{6}, []int{0, 6}}
var worseGroupDiffGoodness = roundGoodness{[]int{3}, []int{1, 2}, []int{6}, []int{0, 6}}
var mildSideDiffsGoodness = roundGoodness{[]int{3}, []int{3}, []int{0, 2, 4}, []int{0, 6}}
var milderSideDiffsGoodness = roundGoodness{[]int{3}, []int{3}, []int{0, 4, 2}, []int{0, 6}}
var mildStreaksGoodness = roundGoodness{[]int{3}, []int{3}, []int{6}, []int{0, 2, 4}}
var milderStreaksGoodness = roundGoodness{[]int{3}, []int{3}, []int{6}, []int{0, 4, 2}}
var badSideDiffsGoodness = roundGoodness{[]int{3}, []int{3}, []int{4, 0, 0, 2}, []int{0, 6}}
var awfulSideDiffsGoodness = roundGoodness{[]int{3}, []int{3}, []int{2, 0, 0, 4}, []int{0, 6}}
var badStreaksGoodness = roundGoodness{[]int{3}, []int{3}, []int{6}, []int{0, 4, 0, 2}}
var awfulStreaksGoodness = roundGoodness{[]int{3}, []int{3}, []int{6}, []int{0, 2, 0, 4}}

// side diffs of one cannot and should not be avoided, so this is just as good as the ideal round
var nearlyIdealRoundGoodness = roundGoodness{[]int{3}, []int{3}, []int{0, 6}, []int{0, 6}}

var goodnessesInOrder = []*roundGoodness{
	&idealRoundGoodness,
	&milderStreaksGoodness,
	&mildStreaksGoodness,
	&milderSideDiffsGoodness,
	&mildSideDiffsGoodness,
	&groupDiffGoodness,
	&worseGroupDiffGoodness,
	&badStreaksGoodness,
	&awfulStreaksGoodness,
	&badSideDiffsGoodness,
	&awfulSideDiffsGoodness,
	&rematchGoodness,
	&multiRematchGoodness,
}

func TestGoodnessComparison(t *testing.T) {
	for i, g1 := range goodnessesInOrder {
		for j, g2 := range goodnessesInOrder {
			e := (j > i)
			r := g1.BetterThan(g2)
			if r != e {
				t.Error("For", g1,
					"better than", g2,
					"expected", e,
					"got", r,
				)
			}
		}
	}

	if idealRoundGoodness.BetterThan(&nearlyIdealRoundGoodness) {
		t.Error("Ideal round compared better than nearly ideal round; expected side diffs of one to be ignored")
	}
	if nearlyIdealRoundGoodness.BetterThan(&idealRoundGoodness) {
		t.Error("Nearly ideal round compared better than ideal round; expected side diffs of one to be ignored")
	}
}
