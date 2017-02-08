package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"math/rand"
	"sort"
)

type Tournament struct {
	Name        string
	Players     []Player
	Standings   []PlayerID
	Rounds      []Round
	SosUpToDate bool
	ScoreGroups map[int]int
}

func (t *Tournament) Player(id PlayerID) *Player {
	if id == NoPlayer || int(id) > len(t.Players) {
		return nil
	}
	return &(t.Players[id-1])
}

func (t *Tournament) AddPlayer(Name string, Corp string, Runner string) error {
	if Name == "" {
		return errors.New("Player name cannot be blank")
	}
	for _, pi := range t.Players {
		if pi.Name == Name {
			return errors.New("Duplicate player name")
		}
	}
	var id PlayerID = PlayerID(len(t.Players) + 1)
	t.Players = append(t.Players, Player{Name: Name, Corp: Corp, Runner: Runner, Tournament: t, PlayerID: id})
	t.Standings = append(t.Standings, id)
	return nil
}

func (t *Tournament) DropPlayer(p PlayerID) {
	t.Player(p).Dropped = true
}

func (t *Tournament) ReAddPlayer(p PlayerID) {
	t.Player(p).Dropped = false
}

func (t *Tournament) NextRound() error {
	if len(t.Rounds) != 0 {
		e := t.Rounds[len(t.Rounds)-1].Finish()
		if e != nil {
			return e
		}
	}

	t.Rounds = append(t.Rounds, Round{Tournament: t, Number: len(t.Rounds) + 1})
	t.Rounds[len(t.Rounds)-1].MakeMatches()
	t.Rounds[len(t.Rounds)-1].Start()
	return nil
}

type Player struct {
	PlayerID
	Tournament      *Tournament `json:"-"`
	Name            string
	Corp            string
	Runner          string
	Prestige        int
	PrestigeAvg     float64
	SoS             float64
	XSoS            float64
	CurrentMatch    MatchID
	FinishedMatches []MatchID
	Dropped         bool
}

type PlayerID int

const NoPlayer PlayerID = -1

// playerSorter joins a Tournament pointer and a slice of PlanetIDs to be sorted.
type playerSorter struct {
	t *Tournament
	p []PlayerID
}

// Len is part of sort.Interface.
func (s *playerSorter) Len() int {
	return len(s.p)
}

// Swap is part of sort.Interface.
func (s *playerSorter) Swap(i, j int) {
	s.p[i], s.p[j] = s.p[j], s.p[i]
}

// Less is part of sort.Interface. It is implemented by calling the "by" closure in the sorter.
func (s *playerSorter) Less(i, j int) bool {
	pi := s.t.Player(s.p[i])
	pj := s.t.Player(s.p[j])

	if pi.Prestige != pj.Prestige {
		return pi.Prestige > pj.Prestige
	} else if pi.SoS != pj.SoS {
		return pi.SoS > pj.SoS
	} else {
		return pi.XSoS > pj.XSoS
	}
}

func (t *Tournament) updateSoS() {
	if !t.SosUpToDate {
		// Update prestige averages
		for i, _ := range t.Players {
			p := &(t.Players[i])
			// Note that byes are counted here, because
			// that's what TOME does
			if len(p.FinishedMatches) == 0 {
				p.PrestigeAvg = 0
			} else {
				p.PrestigeAvg = float64(p.Prestige) / float64(len(p.FinishedMatches))
			}
		}

		// Update SoS
		for i, _ := range t.Players {
			p := &(t.Players[i])
			var SoSSum float64
			var matchCount int
			for _, mID := range p.FinishedMatches {
				m := t.Match(mID)
				if !m.IsBye() {
					SoSSum += t.Player(m.GetOpponent(p.PlayerID)).PrestigeAvg
					matchCount += 1
				}
			}
			if matchCount == 0 {
				p.SoS = 0
			} else {
				p.SoS = SoSSum / float64(matchCount)
			}
		}

		// Update xSoS
		for i, _ := range t.Players {
			p := &(t.Players[i])
			var xSoSSum float64
			var matchCount int
			for _, mID := range p.FinishedMatches {
				m := t.Match(mID)
				if !m.IsBye() {
					xSoSSum += t.Player(m.GetOpponent(p.PlayerID)).SoS
					matchCount += 1
				}
			}
			if matchCount == 0 {
				p.XSoS = 0
			} else {
				p.XSoS = xSoSSum / float64(matchCount)
			}
		}
		t.SosUpToDate = true
	}
}

func (t *Tournament) sortPlayers(p []PlayerID) {
	t.updateSoS()
	t.ScoreGroups = orderPlayers(t, t.Standings, false)
}

func shuffleGroups(t *Tournament, players []PlayerID) {
	orderPlayers(t, players, true)
}

func orderPlayers(t *Tournament, players []PlayerID, shuffleGroups bool) (scoreGroups map[int]int) {
	sort.Sort(&playerSorter{t, players})

	// Record & sort or shuffle the score groups
	scoreGroups = make(map[int]int)
	group := 1
	score := -1
	groupStart := 0
	for i, p := range players {
		if t.Player(p).Prestige != score {
			score = t.Player(p).Prestige
			scoreGroups[score] = group
			group += 1
			if i != 0 {
				if shuffleGroups {
					shufflePlayers(players[groupStart : i-1])
				} else {
					sortScoreGroup(t, players[groupStart:i-1])
				}
			}
			groupStart = i
		}
	}
	// sort last score group
	if shuffleGroups {
		shufflePlayers(players[groupStart:])
	} else {
		sortScoreGroup(t, players[groupStart:])
	}

	return scoreGroups
}

// sortScoreGroup actually just randomizes ties; SoS and xSoS are handled when the whole list is sorted
func sortScoreGroup(t *Tournament, g []PlayerID) {
	tieStart := 0
	SoS := -1.0
	xSoS := -1.0
	for i, p := range g {
		if t.Player(p).SoS != SoS || t.Player(p).XSoS != xSoS {
			SoS = t.Player(p).SoS
			xSoS = t.Player(p).XSoS
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
func shufflePlayers(g []PlayerID) {
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
	Tournament *Tournament `json:"-"`
	Number     int
	Matches    []Match
	Started    bool
	Finished   bool
}

type partialRound struct {
	Tournament       *Tournament
	Pairings         []Pairing
	UnmatchedPlayers []PlayerID
	goodness         roundGoodness
}

type pairingDetails struct {
	rematch     int    // number of times these players have played already
	groupDiff   int    // difference between the group numbers of the two players
	sideDiffs   [2]int // for each player, what their side diff will be after the round
	streaks     [2]int // for each player, what their streak will be after the round
	isBye       bool   // whether this is a bye
	byePrestige int    // how many points the player has, if this is a bye
}

type roundGoodness struct {
	rematches   []int // rematches[i] = number of pairs that are matched for the i-th time
	groupDiffs  []int // groupDiffs[i] = number of pairs that are matched across i groups
	sideDiffs   []int // sideDiffs[i] = number of players that will have a side diff of i after the round
	streaks     []int // streaks[i] = number of players that will have a streak of i after the round
	hasBye      bool  // whether there's at least one player with a bye
	byePrestige int   // how many prestige points the player with the bye has
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

	// better to assign the bye to a player with a lower score
	if g1.hasBye && g2.hasBye {
		if g1.byePrestige < g2.byePrestige {
			return true
		} else if g1.byePrestige > g2.byePrestige {
			return false
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

	if p.isBye {
		if !g.hasBye || p.byePrestige > g.byePrestige {
			// if multiple byes, byePrestige is the highest of the prestiges of players with byes
			g.byePrestige = p.byePrestige
		}
		g.hasBye = true
	}

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
func (t *Tournament) pairingEffects(corpID, runnerID PlayerID) pairingDetails {
	var d pairingDetails

	corp := t.Player(corpID)
	runner := t.Player(runnerID)
	for _, mID := range corp.FinishedMatches {
		m := t.Match(mID)
		if m.Corp == runnerID || m.Runner == runnerID {
			d.rematch += 1
		}
	}

	if runnerID == NoPlayer {
		d.isBye = true
		d.byePrestige = corp.Prestige

		d.groupDiff = 0 // don't include byes in score group difference comparisons

		d.sideDiffs[0], d.streaks[0] = t.playerByeEffects(corpID)
	} else {
		d.groupDiff = t.ScoreGroups[corp.Prestige] - t.ScoreGroups[runner.Prestige]
		if d.groupDiff < 0 {
			d.groupDiff = -d.groupDiff
		}

		d.sideDiffs[0], d.streaks[0] = t.playerCorpEffects(corpID)
		d.sideDiffs[1], d.streaks[1] = t.playerRunnerEffects(runnerID)
	}

	return d
}

func (t *Tournament) playerCorpEffects(p PlayerID) (sideDiff, streak int) {
	sideDiff = 1
	streak = 1
	for _, mID := range t.Player(p).FinishedMatches {
		m := t.Match(mID)
		if !m.IsBye() {
			if m.Corp == p {
				sideDiff += 1
				streak += 1
			} else {
				// runner
				sideDiff -= 1
				streak = 1
			}
		}
	}
	if sideDiff < 0 {
		sideDiff = -sideDiff
	}
	return
}

func (t *Tournament) playerRunnerEffects(p PlayerID) (sideDiff, streak int) {
	sideDiff = -1
	streak = 1
	for _, mID := range t.Player(p).FinishedMatches {
		m := t.Match(mID)
		if !m.IsBye() {
			if m.Corp == p {
				sideDiff += 1
				streak = 1
			} else {
				// runner
				sideDiff -= 1
				streak += 1
			}
		}
	}
	if sideDiff < 0 {
		sideDiff = -sideDiff
	}
	return
}

func (t *Tournament) playerByeEffects(p PlayerID) (sideDiff, streak int) {
	sideDiff = 0
	streak = 0
	runnerStreak := true
	for _, mID := range t.Player(p).FinishedMatches {
		m := t.Match(mID)
		if !m.IsBye() {
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
	}
	if sideDiff < 0 {
		sideDiff = -sideDiff
	}
	return
}

func (p *partialRound) appendMatch(corp, runner PlayerID) partialRound {
	var newPartial partialRound
	newPartial.Pairings = append([]Pairing(nil), p.Pairings...)
	newPartial.Pairings = append(newPartial.Pairings, Pairing{Corp: corp, Runner: runner})

	newPartial.Tournament = p.Tournament

	newPartial.goodness = p.goodness
	newPartial.goodness.addPairing(newPartial.Tournament.pairingEffects(corp, runner))

	if len(p.UnmatchedPlayers) > 2 {
		newPartial.UnmatchedPlayers = make([]PlayerID, 0, len(p.UnmatchedPlayers)-2)

		for _, unmatchedPlayer := range p.UnmatchedPlayers {
			if unmatchedPlayer != corp && unmatchedPlayer != runner {
				newPartial.UnmatchedPlayers = append(newPartial.UnmatchedPlayers, unmatchedPlayer)
			}
		}
	} else {
		newPartial.UnmatchedPlayers = make([]PlayerID, 0)
	}

	return newPartial
}

func (p *partialRound) NextMatches(partials chan partialRound, stop chan int) {
	if len(p.UnmatchedPlayers) == 1 {
		partials <- p.appendMatch(p.UnmatchedPlayers[0], NoPlayer)
	} else if len(p.UnmatchedPlayers) == 2 {
		partials <- p.appendMatch(p.UnmatchedPlayers[0], p.UnmatchedPlayers[1])
		partials <- p.appendMatch(p.UnmatchedPlayers[1], p.UnmatchedPlayers[0])
	} else if len(p.UnmatchedPlayers) > 2 {
		for _, playerB := range p.UnmatchedPlayers[1:] {
			partials <- p.appendMatch(p.UnmatchedPlayers[0], playerB)
			partials <- p.appendMatch(playerB, p.UnmatchedPlayers[0])
		}
		if len(p.UnmatchedPlayers)%2 == 1 {
			partials <- p.appendMatch(p.UnmatchedPlayers[0], NoPlayer)
		}
	}
	stop <- 1
}

func (t Tournament) activePlayers() []PlayerID {
	var players []PlayerID
	for _, p := range t.Players {
		if !p.Dropped {
			players = append(players, p.PlayerID)
		}
	}
	return players
}

func (r *Round) MakeMatches() {
	var bestPairings []Pairing
	if r.Number == 1 {
		players := r.Tournament.activePlayers()
		shufflePlayers(players)
		if len(players)%2 == 1 {
			players = append(players, NoPlayer)
		}

		bestPairings = make([]Pairing, 0, len(players)/2)
		for i := 0; i < len(players)/2; i++ {
			bestPairings = append(bestPairings, Pairing{Corp: players[2*i], Runner: players[2*i+1]})
		}
	} else {
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
					} else {
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
		basePartialMatch.UnmatchedPlayers = r.Tournament.activePlayers()
		shuffleGroups(r.Tournament, basePartialMatch.UnmatchedPlayers)

		basePartialMatch.Tournament = r.Tournament
		partials <- basePartialMatch
		stops <- 1

		bestPairings = <-result
	}
	r.Matches = make([]Match, 0, len(r.Tournament.Players)/2)
	for i, pairing := range bestPairings {
		r.Matches = append(r.Matches, Match{Game: Game{Pairing: pairing}, Number: i + 1})
		if pairing.Runner == NoPlayer {
			// bye
			r.Matches[len(r.Matches)-1].Game.RecordResult(pairing.Corp, false)
		}
	}
}

func (r *Round) Start() {
	if !r.Started {
		r.Started = true
		for _, m := range r.Matches {
			mID := MatchID{r.Number, m.Number}
			r.Tournament.Player(m.Corp).CurrentMatch = mID
			if m.Runner != NoPlayer {
				r.Tournament.Player(m.Runner).CurrentMatch = mID
			}
		}
	}
}

func (r *Round) Finish() error {
	if r.Started && !r.Finished {
		r.Finished = true
		for _, m := range r.Matches {
			if !m.Concluded {
				return errors.New("Some matches not recorded")
			}
		}
		for _, m := range r.Matches {
			mID := MatchID{r.Number, m.Number}
			corp := r.Tournament.Player(m.Corp)
			runner := r.Tournament.Player(m.Runner)
			corp.Prestige += m.GetPrestige(m.Corp)
			corp.FinishedMatches = append(corp.FinishedMatches, mID)
			corp.CurrentMatch = MatchID{}
			if runner != nil {
				runner.Prestige += m.GetPrestige(m.Runner)
				runner.FinishedMatches = append(runner.FinishedMatches, mID)
				runner.CurrentMatch = MatchID{}
			}
		}

		r.Tournament.SosUpToDate = false
		r.Tournament.updateSoS()
		r.Tournament.sortPlayers(r.Tournament.Standings)
	}

	return nil
}

func (g *Game) RecordResult(winner PlayerID, modifiedWin bool) {
	g.Concluded = true
	g.CorpWin = (winner == g.Corp)
	g.RunnerWin = (winner == g.Runner)
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
	Corp   PlayerID
	Runner PlayerID
}

type Match struct {
	Game
	Number int
}

type MatchID struct {
	Round int
	Match int
}

func (t *Tournament) Match(m MatchID) *Match {
	if m.Round < 1 || m.Round > len(t.Rounds) {
		return nil
	}

	r := &(t.Rounds[m.Round-1])

	if m.Match < 1 || m.Match > len(r.Matches) {
		return nil
	}

	return &(r.Matches[m.Match-1])
}

func (m Match) IsBye() bool {
	return (m.Corp == NoPlayer || m.Runner == NoPlayer)
}
func (m Match) IsDone() bool {
	return m.Game.Concluded
}
func (m Match) GetPrestige(p PlayerID) int {
	if p != m.Corp && p != m.Runner {
		return 0
	} else if m.Runner == NoPlayer || m.Corp == NoPlayer {
		//Bye
		return 3
	} else if m.Corp == p {
		return m.Game.CorpPrestige()
	} else {
		return m.Game.RunnerPrestige()
	}
}
func (m Match) GetOpponent(p PlayerID) PlayerID {
	if p == m.Corp {
		return m.Runner
	} else if p == m.Runner {
		return m.Corp
	} else {
		return NoPlayer
	}
}

func (m Match) GetWinner() PlayerID {
	if m.Game.RunnerWin {
		return m.Runner
	} else if m.Game.CorpWin {
		return m.Corp
	} else {
		return NoPlayer
	}
}

func (t *Tournament) save(file string) error {
	marshaled, e := json.Marshal(t)
	if e != nil {
		return e
	}
	return ioutil.WriteFile(file, marshaled, 0600)
}

func loadTournament(t *Tournament, file string) error {
	marshaled, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	err = json.Unmarshal(marshaled, t)
	if err != nil {
		return err
	}
	for i, _ := range t.Players {
		t.Players[i].Tournament = t
	}
	for i, _ := range t.Rounds {
		t.Rounds[i].Tournament = t
	}
	return nil
}
