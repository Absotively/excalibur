package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
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
	t.Rounds = append(t.Rounds, Round{Tournament: t, Number: len(t.Rounds) + 1})
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

func (t *Tournament) updateSoS() {
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
}

func (t *Tournament) SortPlayers() {
	t.updateSoS()
	sort.Sort(t.Players)

	// Record the score groups
	group := 1
	score := -1
	groupStart := 0
	for i, p := range t.Players {
		if p.Prestige != score {
			score = p.Prestige
			t.ScoreGroups[score] = group
			group += 1
			if i != 0 {
				sortScoreGroup(t.Players[groupStart : i-1])
			}
			groupStart = i
		}
	}
	// sort last score group
	sortScoreGroup(t.Players[groupStart:])
}

// sortScoreGroup actually just randomizes ties; SoS and xSoS are handled when the whole list is sorted
func sortScoreGroup(g Players) {
	tieStart := 0
	SoS := -1.0
	xSoS := -1.0
	for i, p := range g {
		if p.SoS != SoS || p.XSoS != xSoS {
			SoS = p.SoS
			xSoS = p.XSoS
			if i != 0 && i-tieStart > 1 {
				shufflePlayers(g[tieStart : i-1])
			}
			tieStart = i
		}
	}
	// shuffle last tie group
	shufflePlayers(g[tieStart:])
}

// basically copied from http://marcelom.github.io/2013/06/07/goshuffle.html
func shufflePlayers(g Players) {
	for i := range g {
		j := rand.Intn(i + 1)
		g[i], g[j] = g[j], g[i]
	}
}

type Game struct {
	Pairing
	Concluded   bool
	CorpWin     bool
	RunnerWin   bool
	ModifiedWin bool
}

type Round struct {
	Tournament *Tournament
	Number     int
	Matches    []Match
}

type partialRound struct {
	Tournament       *Tournament
	Pairings         []Pairing
	UnmatchedPlayers []*Player
	goodness         roundGoodness
}

type pairingDetails struct {
	rematch   int    // number of times these players have played already
	groupDiff int    // difference between the group numbers of the two players
	sideDiffs [2]int // for each player, what their side diff will be after the round
	streaks   [2]int // for each player, what their streak will be after the round
}

type roundGoodness struct {
	rematches  []int // rematches[i] = number of pairs that are matched for the i-th time
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
func (t *Tournament) PairingEffects(corp, runner *Player) pairingDetails {
	var d pairingDetails

	for _, m := range corp.FinishedMatches {
		if m.Corp == runner || m.Runner == runner {
			d.rematch += 1
		}
	}

	if runner == nil {
		var lowestScore int
		for s := range t.ScoreGroups {
			if s < lowestScore {
				lowestScore = s
			}
		}

		d.groupDiff = t.ScoreGroups[corp.Prestige] - t.ScoreGroups[lowestScore]
		d.groupDiff += 1 // treat bye as being in its own score group below the last score group

		d.sideDiffs[0], d.streaks[0] = playerByeEffects(corp)
	} else {
		d.groupDiff = t.ScoreGroups[corp.Prestige] - t.ScoreGroups[runner.Prestige]
		if d.groupDiff < 0 {
			d.groupDiff = -d.groupDiff
		}

		d.sideDiffs[0], d.streaks[0] = playerCorpEffects(corp)
		d.sideDiffs[1], d.streaks[1] = playerRunnerEffects(runner)
	}

	return d
}

func playerCorpEffects(p *Player) (sideDiff, streak int) {
	sideDiff = 1
	streak = 1
	for _, m := range p.FinishedMatches {
		if m.Corp == p {
			sideDiff += 1
			streak += 1
		} else {
			// runner
			sideDiff -= 1
			streak = 1
		}
	}
	return
}

func playerRunnerEffects(p *Player) (sideDiff, streak int) {
	sideDiff = -1
	streak = 1
	for _, m := range p.FinishedMatches {
		if m.Corp == p {
			sideDiff += 1
			streak = 1
		} else {
			// runner
			sideDiff -= 1
			streak += 1
		}
	}
	return
}

func playerByeEffects(p *Player) (sideDiff, streak int) {
	sideDiff = 0
	streak = 0
	runnerStreak := true
	for _, m := range p.FinishedMatches {
		if m.Corp == p {
			sideDiff += 1
			if runnerStreak {
				streak = 1
				runnerStreak = false
			} else {
				streak += 1
			}
		} else {
			// runner
			sideDiff -= 1
			if runnerStreak {
				streak += 1
			} else {
				streak = 1
				runnerStreak = true
			}
		}
	}
	return
}

func (p *partialRound) appendMatch(corp, runner *Player) partialRound {
	var newPartial partialRound
	newPartial.Pairings = append([]Pairing(nil), p.Pairings...)
	newPartial.Pairings = append(newPartial.Pairings, Pairing{Corp: corp, Runner: runner})

	newPartial.Tournament = p.Tournament

	newPartial.goodness = p.goodness
	newPartial.goodness.addPairing(newPartial.Tournament.PairingEffects(corp, runner))

	if len(p.UnmatchedPlayers) > 2 {
		newPartial.UnmatchedPlayers = make([]*Player, len(p.UnmatchedPlayers)-2)

		for _, unmatchedPlayer := range p.UnmatchedPlayers {
			if unmatchedPlayer != corp && unmatchedPlayer != runner {
				newPartial.UnmatchedPlayers = append(newPartial.UnmatchedPlayers, unmatchedPlayer)
			}
		}
	} else {
		newPartial.UnmatchedPlayers = make([]*Player, 0)
	}

	return newPartial
}

func (p *partialRound) NextMatches(partials chan partialRound, stop chan int) {
	if len(p.UnmatchedPlayers) == 1 {
		partials <- p.appendMatch(p.UnmatchedPlayers[0], nil)
	} else if len(p.UnmatchedPlayers) == 2 {
		partials <- p.appendMatch(p.UnmatchedPlayers[0], p.UnmatchedPlayers[1])
		partials <- p.appendMatch(p.UnmatchedPlayers[1], p.UnmatchedPlayers[0])
	} else if len(p.UnmatchedPlayers) > 2 {
		for _, playerB := range p.UnmatchedPlayers[1:] {
			partials <- p.appendMatch(p.UnmatchedPlayers[0], playerB)
			partials <- p.appendMatch(playerB, p.UnmatchedPlayers[0])
		}
		if len(p.UnmatchedPlayers)%2 == 1 {
			partials <- p.appendMatch(p.UnmatchedPlayers[0], nil)
		}
	}
	stop <- 1
}

func (r *Round) MakeMatches() {
	var bestPairings []Pairing
	if r.Number == 1 {
		players := append([]*Player(nil), r.Tournament.Players...)
		shufflePlayers(players)
		if len(players)%2 == 1 {
			players = append(players, nil)
		}

		bestPairings = make([]Pairing, len(players)/2)
		for i := 0; i < len(players)/2; i++ {
			bestPairings = append(bestPairings, Pairing{Corp: players[2*i], Runner: players[2*i+1]})
		}
	} else {
		r.Tournament.SortPlayers()
		partials := make(chan partialRound)
		stops := make(chan int)
		result := make(chan []Pairing)
		go func(partials chan partialRound, stops chan int) {
			activeRoundCreators := 1 // MakeMatches() is one until it's sent off the base partial
			var bestSoFar partialRound
			for {
				select {
				case p := <-partials:
					// further matching can only make derivatives of p worse,
					// so if it's not currently better than bestSoFar then
					// it never will be and we can just toss it out now
					if len(bestSoFar.Pairings) == 0 || p.goodness.BetterThan(&(bestSoFar.goodness)) {
						if len(p.UnmatchedPlayers) == 0 {
							bestSoFar = p
						} else {
							go p.NextMatches(partials, stops)
							activeRoundCreators += 1
						}
					}
				case <-stops:
					activeRoundCreators -= 1
				default:
					if activeRoundCreators == 0 {
						result <- bestSoFar.Pairings
						return
					}
				}
			}
		}(partials, stops)

		var basePartialMatch partialRound
		copy(basePartialMatch.UnmatchedPlayers, r.Tournament.Players)
		basePartialMatch.Tournament = r.Tournament
		shufflePlayers(basePartialMatch.UnmatchedPlayers)
		partials <- basePartialMatch
		stops <- 1

		bestPairings = <-result
	}
	r.Matches = make([]Match, len(r.Tournament.Players)/2)
	for _, pairing := range bestPairings {
		r.Matches = append(r.Matches, Match{Game: Game{Pairing: pairing}})
	}
}

func (r *Round) Start() {
	for _, m := range r.Matches {
		m.Corp.CurrentMatch = &m
		m.Runner.CurrentMatch = &m
	}
}

func (r *Round) Finish() {
	for _, m := range r.Matches {
		m.Corp.Prestige += m.GetPrestige(m.Corp)
		m.Runner.Prestige += m.GetPrestige(m.Runner)

		m.Corp.FinishedMatches = append(m.Corp.FinishedMatches, &m)
		m.Runner.FinishedMatches = append(m.Runner.FinishedMatches, &m)

		m.Corp.CurrentMatch = nil
		m.Runner.CurrentMatch = nil
	}

	r.Tournament.updateSoS()
}

func (g *Game) RecordResult(winner *Player, modifiedWin bool) {
	g.Concluded = true
	if winner == g.Corp {
		g.CorpWin = true
		g.RunnerWin = false
	} else if winner == g.Runner {
		g.RunnerWin = true
		g.CorpWin = false
	}
	g.ModifiedWin = modifiedWin
}

func (g Game) CorpPrestige() int {
	if !g.Concluded {
		return 0
	} else if g.CorpWin {
		if g.ModifiedWin {
			return 2
		} else {
			return 3
		}
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
		if g.ModifiedWin {
			return 2
		} else {
			return 3
		}
	} else if g.CorpWin {
		return 0
	} else {
		return 1
	}
}

type Pairing struct {
	Corp   *Player
	Runner *Player
}

type Match struct {
	Game
}

func (m Match) IsBye() bool {
	return (m.Corp == nil || m.Runner == nil)
}
func (m Match) IsDone() bool {
	return m.Game.Concluded
}
func (m Match) GetPrestige(p *Player) int {
	if p != m.Corp && p != m.Runner {
		return 0
	} else if m.Runner == nil || m.Corp == nil {
		//Bye
		return 2
	} else if m.Corp == p {
		return m.Game.CorpPrestige()
	} else {
		return m.Game.RunnerPrestige()
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

func (m Match) GetWinner() *Player {
	if m.Game.RunnerWin {
		return m.Runner
	} else if m.Game.CorpWin {
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
