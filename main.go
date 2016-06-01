package main

import (
	"fmt"
	"io/ioutil"
	"sort"
)

type Tournament struct {
	Name string
	Players
	Rounds      []Round
	sosUpToDate bool
	ScoreGroups map[int]int
}

func (t *Tournament) AddPlayer(Name string, Corp string, Runner string) {
	t.Players = append(t.Players, &Player{Name: Name, Corp: Corp, Runner: Runner})
}

func (t *Tournament) AddRound() {
	t.Rounds = append(t.Rounds, Round{})
}

type Player struct {
	Name            string
	Corp            string
	Runner          string
	Prestige        int
	PrestigeAvg     float64
	SoS             float64
	XSoS            float64
	CurrentMatch    *Match
	FinishedMatches []*Match
}

type Players []*Player

func (s Players) Len() int      { return len(s) }
func (s Players) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s Players) Less(i, j int) bool {
	if s[i].Prestige != s[j].Prestige {
		return s[i].Prestige < s[j].Prestige
	} else if s[i].SoS != s[j].SoS {
		return s[i].SoS < s[j].SoS
	} else {
		return s[i].XSoS < s[j].XSoS
	}
}

func (t *Tournament) SortPlayers() {
	if !t.sosUpToDate {
		// Update prestige averages
		for _, p := range t.Players {
			// Note that byes are counted here, because
			// that's what TOME does
			if len(p.FinishedMatches) == 0 {
				p.PrestigeAvg = 0
			} else {
				p.PrestigeAvg = float64(p.Prestige) / float64(len(p.FinishedMatches))
			}
		}

		// Update SoS
		for _, p := range t.Players {
			var SoSSum int
			var matchCount int
			for _, m := range p.FinishedMatches {
				if !m.IsBye() {
					SoSSum += m.GetOpponent(p).Prestige
					matchCount += 1
				}
			}
			p.SoS = float64(SoSSum) / float64(matchCount)
		}

		// Update xSoS
		for _, p := range t.Players {
			var xSoSSum float64
			var matchCount int
			for _, m := range p.FinishedMatches {
				if !m.IsBye() {
					xSoSSum += m.GetOpponent(p).XSoS
					matchCount += 1
				}
			}
			p.XSoS = xSoSSum / float64(matchCount)
		}
		t.sosUpToDate = true
	}

	sort.Sort(t.Players)
	// TODO: Handle h2h, randomize ties

	// Record the score groups
	group := 1
	score := -1
	for _, p := range t.Players {
		if p.Prestige != score {
			score = p.Prestige
			t.ScoreGroups[score] = group
			group += 1
		}
	}
}

type Game struct {
	Corp      *Player
	Runner    *Player
	Concluded bool
	CorpWin   bool
	RunnerWin bool
}

type Round struct {
	Tournament *Tournament
	Number     int
	Matches    []*Match
}

type partialRound struct {
	Pairings         [][2]*Player
	UnmatchedPlayers []*Player
	GroupCrossings   []int
	SideDifferences  []int
	SideStreaks      []int
}

type pairingDetails struct {
	rematch   int    // number of times these players have played already
	groupDiff int    // difference between the group numbers of the two players
	sideDiffs [2]int // for each player, what their side diff will be after the round
	streaks   [2]int // for each player, what their streak will be after the round
}

type roundGoodness struct {
	rematches  []int // rematches[i] = number of pairs that are rematched for the i-th time
	groupDiffs []int // groupDiffs[i] = number of pairs that are matched across i groups
	sideDiffs  []int // sideDiffs[i] = number of players that will have a side diff of i after the round
	streaks    []int // streaks[i] = number of players that will have a streak of i after the round
}

func (g1 *roundGoodness) BetterThan(g2 *roundGoodness) bool {
	// rematches bad
	if len(g1.rematches) < len(g2.rematches) {
		return true
	} else if len(g1.rematches) > len(g2.rematches) {
		return false
	} else if len(g1.rematches) > 1 {
		for i := len(g1.rematches) - 1; i > 0; i-- {
			if g1.rematches[i] < g2.rematches[i] {
				return true
			} else if g1.rematches[i] > g2.rematches[i] {
				return false
			}
		}
	}

	// side diffs of more than two bad
	if len(g1.sideDiffs) > 3 || len(g2.sideDiffs) > 3 {
		if len(g1.sideDiffs) < len(g2.sideDiffs) {
			return true
		} else if len(g1.sideDiffs) > len(g2.sideDiffs) {
			return false
		} else {
			for i := len(g1.sideDiffs) - 1; i > 2; i-- {
				if g1.sideDiffs[i] < g2.sideDiffs[i] {
					return true
				} else if g1.sideDiffs[i] > g2.sideDiffs[i] {
					return false
				}
			}
		}
	}

	// streaks of more than two bad
	if len(g1.streaks) > 3 || len(g2.streaks) > 3 {
		if len(g1.streaks) < len(g2.streaks) {
			return true
		} else if len(g1.streaks) > len(g2.streaks) {
			return false
		} else {
			for i := len(g1.streaks) - 1; i > 2; i-- {
				if g1.streaks[i] < g2.streaks[i] {
					return true
				} else if g1.streaks[i] > g2.streaks[i] {
					return false
				}
			}
		}
	}

	// minimize pairings that cross score groups
	if len(g1.groupDiffs) < len(g2.groupDiffs) {
		return true
	} else if len(g1.groupDiffs) > len(g2.groupDiffs) {
		return false
	} else {
		for i := len(g1.groupDiffs) - 1; i >= 0; i-- {
			if g1.groupDiffs[i] < g2.groupDiffs[i] {
				return true
			} else if g1.groupDiffs[i] > g2.groupDiffs[i] {
				return false
			}
		}
	}

	// side diffs of two are mildly undesirable
	if len(g1.sideDiffs) > 2 {
		if len(g2.sideDiffs) <= 2 {
			return false
		} else if g1.sideDiffs[2] < g2.sideDiffs[2] {
			return true
		} else if g1.sideDiffs[2] > g2.sideDiffs[2] {
			return false
		}
	} else if len(g2.sideDiffs) > 2 {
		return true
	}

	// streaks of two are mildly undesirable
	if len(g1.streaks) > 2 {
		if len(g2.streaks) <= 2 {
			return false
		} else if g1.streaks[2] < g2.streaks[2] {
			return true
		} else if g1.streaks[2] > g2.streaks[2] {
			return false
		}
	} else if len(g2.streaks) > 2 {
		return true
	}

	return false
}

func (g *roundGoodness) addPairing(p pairingDetails) {
	if len(g.rematches)-1 < p.rematch {
		a := p.rematch - len(g.rematches) + 1
		g.rematches = append(g.rematches, make([]int, a)...)
	}
	g.rematches[p.rematch] += 1

	if len(g.groupDiffs)-1 < p.groupDiff {
		a := p.groupDiff - len(g.groupDiffs) + 1
		g.groupDiffs = append(g.groupDiffs, make([]int, a)...)
	}
	g.groupDiffs[p.groupDiff] += 1

	m := p.sideDiffs[0]
	if p.sideDiffs[1] > m {
		m = p.sideDiffs[1]
	}
	if len(g.sideDiffs)-1 < m {
		a := m - len(g.sideDiffs) + 1
		g.sideDiffs = append(g.sideDiffs, make([]int, a)...)
	}
	g.sideDiffs[p.sideDiffs[0]] += 1
	g.sideDiffs[p.sideDiffs[1]] += 1

	m = p.streaks[0]
	if p.streaks[1] > m {
		m = p.streaks[1]
	}
	if len(g.streaks)-1 < m {
		a := m - len(g.streaks) + 1
		g.streaks = append(g.streaks, make([]int, a)...)
	}
	g.streaks[p.streaks[0]] += 1
	g.streaks[p.streaks[1]] += 1
}

// will need to cache these at some point, probably
func (t *Tournament) PairingEffects(a, b *Player) pairingDetails {
	var d pairingDetails
	d.groupDiff = t.ScoreGroups[a.Prestige] - t.ScoreGroups[b.Prestige]
	if d.groupDiff < 0 {
		d.groupDiff = -d.groupDiff
	}

	sideDiffA := 0
	streakA := 0
	lastSideA := 0
	rematch := 0
	for _, m := range a.FinishedMatches {
		if m.Corp == a {
			sideDiffA += 1
			if lastSideA == 1 {
				streakA += 1
			} else {
				streakA = 1
				lastSideA = 1
			}
		} else {
			// runner
			sideDiffA -= 1
			if lastSideA == -1 {
				streakA += 1
			} else {
				streakA = 1
				lastSideA = -1
			}
		}
		if (m.Corp == a && m.Runner == b) || (m.Runner == a && m.Corp == b) {
			rematch += 1
		}
	}
	sideDiffB := 0
	streakB := 0
	lastSideB := 0
	for _, m := range b.FinishedMatches {
		if m.Corp == b {
			sideDiffB += 1
			if lastSideB == 1 {
				streakB += 1
			} else {
				streakB = 1
				lastSideB = 1
			}
		} else {
			// runner
			sideDiffB -= 1
			if lastSideB == -1 {
				streakB += 1
			} else {
				streakB = 1
				lastSideB = -1
			}
		}
	}
	d.sideDiffs[0] = sideDiffA
	d.sideDiffs[1] = sideDiffB
	d.streaks[0] = streakA
	d.streaks[1] = streakB
	d.rematch = rematch

	return d
}

func (p *partialRound) NextMatches(partials chan partialRound, stop chan int) {
	if len(p.UnmatchedPlayers) == 1 {
		var newPartial partialRound
		newPartial.Pairings = append(p.Pairings, [2]*Player{p.UnmatchedPlayers[0], nil})
		partials <- newPartial
	} else if len(p.UnmatchedPlayers) == 2 {
		var newPartial partialRound
		newPartial.Pairings = append(p.Pairings, [2]*Player{p.UnmatchedPlayers[0], p.UnmatchedPlayers[1]})
		partials <- newPartial
	} else if len(p.UnmatchedPlayers) > 2 {
		for _, playerB := range p.UnmatchedPlayers[1:] {
			var newPartial partialRound
			newPartial.Pairings = append(p.Pairings, [2]*Player{p.UnmatchedPlayers[0], playerB})
			for _, unmatchedPlayer := range p.UnmatchedPlayers[1:] {
				if unmatchedPlayer != playerB {
					newPartial.UnmatchedPlayers = append(newPartial.UnmatchedPlayers, unmatchedPlayer)
				}
			}
			partials <- newPartial
		}
	}
	stop <- 1
}

func (r *Round) MakeMatches() {
	r.Tournament.SortPlayers()
	partials := make(chan partialRound)
	stops := make(chan int)
	result := make(chan [][2]*Player)
	go func(partials chan partialRound, stops chan int) {
		activeRoundCreators := 1 // One will be started by MakeMatches()
		var bestSoFar [][2]*Player
		for {
			select {
			case p := <-partials:
				if len(p.UnmatchedPlayers) == 0 {
					copy(bestSoFar, p.Pairings)
				} else {
					go p.NextMatches(partials, stops)
					activeRoundCreators += 1
				}
			case <-stops:
				activeRoundCreators -= 1
			default:
				if activeRoundCreators == 0 {
					result <- bestSoFar
					return
				}
			}
		}
	}(partials, stops)
	var basePartialMatch partialRound
	copy(basePartialMatch.UnmatchedPlayers, r.Tournament.Players)
	// TODO: Randomize players within score groups
	partials <- basePartialMatch
	stops <- 1
	//bestPairings <- result
	// TODO: Initialize matches for best pairings
}

func (r *Round) Start() {
	for _, m := range r.Matches {
		m.Corp.CurrentMatch = m
		m.Runner.CurrentMatch = m
	}
}

func (r *Round) Finish() {
	for _, m := range r.Matches {
		m.Corp.Prestige += m.GetPrestige(m.Corp)
		m.Runner.Prestige += m.GetPrestige(m.Runner)

		m.Corp.FinishedMatches = append(m.Corp.FinishedMatches, m)
		m.Runner.FinishedMatches = append(m.Runner.FinishedMatches, m)

		m.Corp.CurrentMatch = nil
		m.Runner.CurrentMatch = nil
	}

	// Update SoS
	for _, p := range r.Tournament.Players {
		var SoSSum int
		var matchCount int
		for _, m := range p.FinishedMatches {
			if !m.IsBye() {
				SoSSum += m.GetOpponent(p).Prestige
				matchCount += 1
			}
		}
		p.SoS = float64(SoSSum) / float64(matchCount)
	}

	// Update xSoS
	for _, p := range r.Tournament.Players {
		var xSoSSum float64
		var matchCount int
		for _, m := range p.FinishedMatches {
			if !m.IsBye() {
				xSoSSum += m.GetOpponent(p).XSoS
				matchCount += 1
			}
		}
		p.XSoS = xSoSSum / float64(matchCount)
	}
}

func (g Game) CorpPrestige() int {
	if !g.Concluded {
		return 0
	} else if g.CorpWin {
		return 2
	} else if g.RunnerWin {
		return 0
	} else {
		return 1
	}
}

func (g Game) RunnerPrestige() int {
	if !g.Concluded {
		return 0
	} else if g.RunnerWin {
		return 2
	} else if g.CorpWin {
		return 0
	} else {
		return 1
	}
}

type Match struct {
	Corp   *Player
	Runner *Player
	G      Game
}

func (m Match) IsBye() bool {
	return (m.Corp == nil || m.Runner == nil)
}
func (m Match) IsDone() bool {
	return m.G.Concluded
}
func (m Match) GetPrestige(p *Player) int {
	if p != m.Corp && p != m.Runner {
		return 0
	} else if m.Runner == nil || m.Corp == nil {
		//Bye
		return 2
	} else if m.Corp == p {
		return m.G.CorpPrestige()
	} else {
		return m.G.RunnerPrestige()
	}
}
func (m Match) GetOpponent(p *Player) *Player {
	if p == m.Corp {
		return m.Runner
	} else if p == m.Runner {
		return m.Corp
	} else {
		return nil
	}
}

func (t *Tournament) save() error {
	filename := t.Name + ".tournament"
	return ioutil.WriteFile(filename, nil, 0600)
}

func loadTournament(title string) (*Tournament, error) {
	filename := title + ".tournament"
	_, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Tournament{Name: title}, nil
}

func main() {
	t1 := &Tournament{Name: "regional"}
	t1.save()
	t2, _ := loadTournament("regional")
	fmt.Println(t2.Name)
}
