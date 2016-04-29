package main

import (
	"fmt"
	"io/ioutil"
	"sort"
)

type Tournament struct {
	Name        string
	Players     Players
	Rounds      []Round
	sosUpToDate boolean
	ScoreGroups map[int]int
}

func (t *Tournament) AddPlayer(Name string, Corp string, Runner string) {
	Players = append(Players, Player{Name: Name, Corp: Corp, Runner: Runner})
}

func (t *Tournament) AddRound() {
	Rounds = append(Rounds, new(Round))
}

type Player struct {
	Name            string
	Corp            string
	Runner          string
	Prestige        int
	PrestigeAvg     float
	SoS             float
	XSoS            float
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
		return s[i].xSoS < s[j].SoS
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
				p.PrestigeAvg = p.Prestige / len(p.FinishedMatches)
			}
		}

		// Update SoS
		for _, p := range t.Players {
			var SoS float
			var matchCount int
			for _, m := range p.FinishedMatches {
				if !m.IsBye {
					SoS += m.GetOpponent(p).Prestige
					matchCount += 1
				}
			}
			p.SoS = SoS / matchCount
		}

		// Update xSoS
		for _, p := range t.Players {
			var xSoS float
			var matchCount int
			for _, m := range p.FinishedMatches {
				if !m.IsBye {
					xSoS += m.GetOpponent(p).XSoS
					matchCount += 1
				}
			}
			p.XSoS = xSoS / matchCount
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
	Concluded boolean
	CorpWin   boolean
	RunnerWin boolean
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

// will need to cache these at some point, probably
func (t *Tournament) PairingEffects(a, b *Player) (rematch int, groupDiff int, sideDiffs [2]int, streaks [2]int) {
	groupDiff = t.ScoreGroups[a.Prestige] - t.ScoreGroups[b.Prestige]
	if groupDiff < 0 {
		groupDiff = -groupDiff
	}

	sideDiffA := 0
	streakA := 0
	lastSideA := 0
	rematch = 0
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
}

func (p *partialRound) NextMatches(partials chan partialRound, stop chan int) {
	if len(p.UnmatchedPlayers) == 1 {
		var newPartial partialRound
		newPartial.Pairings = append(p.Pairings, [2]*Players{&p.UnmatchedPlayers[0], nil})
		partials <- newPartial
	} else if len(p.UnmatchedPlayers) == 2 {
		var newPartial partialRound
		newPartial.Pairings = append(p.Pairings, [2]*Players{&p.UnmatchedPlayers[0], &p.UnmatchedPlayers[1]})
		partials <- newPartial
	} else if len(p.UnmatchedPlayers) > 2 {
		for _, playerB := range p.UnmatchedPlayers[1:] {
			var newPartial partialRound
			newPartial.Pairings = append(p.Pairings, [2]*Players{&p.UnmatchedPlayers[0], playerB})
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
			case partial <- partials:
				if len(partial.UnmatchedPlayers) == 0 {
					copy(bestSoFar, partial.Pairings)
				} else {
					go partial.NextMatches(partials, stops)
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
	}()
	var basePartialMatch partialMatch
	copy(basePartialMatch.UnmatchedPlayers, r.Tournament.Players)
	// TODO: Randomize players within score groups
	partials <- basePartialMatch
	stops <- 1
	bestPairings <- result
	// TODO: Initialize matches for best pairings
}

func (r *Round) Start() {
	for _, m := range r.Matches {
		m.Corp.CurrentMatch = &m
		m.Runner.CurrentMatch = &m
	}
}

func (r *Round) Finish() {
	for _, m := range r.Matches {
		m.PlayerA.Prestige += m.PlayerAPrestige()
		m.PlayerB.Prestige += m.PlayerBPrestige()

		m.PlayerA.FinishedMatches = append(m.PlayerA.FinishedMatches, &m)
		m.PlayerB.FinishedMatches = append(m.PlayerB.FinishedMatches, &m)

		m.PlayerA.CurrentMatch = nil
		m.PlayerB.CurrentMatch = nil
	}

	// Update SoS
	for _, p := range r.Tournament.Players {
		var SoS float
		var matchCount int
		for _, m := range p.FinishedMatches {
			if !m.IsBye {
				SoS += m.GetOpponent(p).Prestige
				matchCount += 1
			}
		}
		p.SoS = SoS / matchCount
	}

	// Update xSoS
	for _, p := range r.Tournament.Players {
		var xSoS float
		var matchCount int
		for _, m := range p.FinishedMatches {
			if !m.IsBye {
				xSoS += m.GetOpponent(p).XSoS
				matchCount += 1
			}
		}
		p.XSoS = xSoS / matchCount
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

func (m Match) IsBye() boolean {
	return (PlayerB == nil)
}
func (m Match) IsDone() boolean {
	return m.G.Completed
}
func (m Match) GetPrestige(p *Player) int {
	if p != Corp && p != Runner {
		return 0
	} else if Runner == nil {
		//Bye
		return 2
	} else if m.Corp == a {
		return m.G.CorpPrestige()
	} else {
		return m.G.RunnerPrestige()
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
