package main

import (
	"errors"
	"fmt"
	"html/template"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

var tournament Tournament

func applyTemplate(w http.ResponseWriter, src string, data interface{}) error {
	t, e := template.New("base").Parse(frameTemplate)
	if e != nil {
		fmt.Println(e.Error())
		return e
	}
	_, e = t.New("content").Parse(src)
	if e != nil {
		fmt.Println(e.Error())
		return e
	}
	e = t.Execute(w, data)
	if e != nil {
		fmt.Println(e.Error())
		return e
	}
	return nil
}

func playerList(w http.ResponseWriter, r *http.Request) {
	applyTemplate(w, playerListTemplate, tournament)
}

func standings(w http.ResponseWriter, r *http.Request) {
	applyTemplate(w, standingsTemplate, tournament)
}

func playerForm(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{"saveurl": r.URL.Path}
	edit := (r.FormValue("edit") != "" || (r.FormValue("name") != "" && r.Method == "GET"))
	if !edit {
		data["add"] = "add"
	}

	var e error
	name := r.FormValue("name")
	oldName := r.FormValue("old-name")
	corp := r.FormValue("corp")
	runner := r.FormValue("runner")

	if r.Method == "POST" {
		if edit {
			e = errors.New("No such player")
			for _, player := range tournament.Players {
				if player.Name == oldName {
					player.Name = name
					player.Corp = corp
					player.Runner = runner
					e = nil
					break
				}
			}
		} else {
			e = tournament.AddPlayer(name, corp, runner)
		}

		if e == nil {
			seeOther(w, "/players")
			return
		}
	}

	// either need initial form or edit/add failed
	if edit {
		if oldName == "" {
			oldName = name
		}
		for _, player := range tournament.Players {
			if player.Name == oldName {
				if name == "" {
					name = player.Name
				}
				if corp == "" {
					corp = player.Corp
				}
				if runner == "" {
					runner = player.Runner
				}
			}
		}
	}

	if e != nil {
		data["error"] = e.Error()
	}
	data["name"] = name
	data["oldName"] = oldName
	data["corp"] = corp
	data["runner"] = runner
	if !edit {
		data["add"] = "add"
	}

	applyTemplate(w, playerFormTemplate, data)
}

func changePlayer(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")

	if name != "" && (r.Method == "GET" || r.FormValue("edit") != "") {
		playerForm(w, r)
		return
	}

	if r.Method != "POST" {
		seeOther(w, "/players")
		return
	}

	for _, player := range tournament.Players {
		if player.Name == name {
			if r.FormValue("drop") != "" {
				tournament.DropPlayer(player)
			} else if r.FormValue("re-add") != "" {
				tournament.ReAddPlayer(player)
			}
			break
		}
	}
	seeOther(w, "/players")
}

func menu(w http.ResponseWriter, r *http.Request) {
	applyTemplate(w, menuTemplate, nil)
}

func startRound(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		e := tournament.NextRound()
		if e != nil {
			applyTemplate(w, errorTemplate, e)
		} else {
			seeOther(w, "/matches")
		}
	} else {
		seeOther(w, "/")
	}
}

func finishRound(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" && len(tournament.Rounds) > 0 {
		e := tournament.Rounds[len(tournament.Rounds)-1].Finish()
		if e != nil {
			applyTemplate(w, errorTemplate, e)
			return
		}
	}
	seeOther(w, "/")
}

func matches(w http.ResponseWriter, r *http.Request) {
	if len(tournament.Rounds) == 0 {
		applyTemplate(w, noMatchesTemplate, Round{})
	} else {
		applyTemplate(w, matchesTemplate, tournament.Rounds[len(tournament.Rounds)-1])
	}
}

func rounds(w http.ResponseWriter, r *http.Request) {
	t, e := template.New("base").Parse(frameTemplate)
	if e != nil {
		fmt.Println(e.Error())
	}
	c, e := t.New("content").Parse(roundsTemplate)
	if e != nil {
		fmt.Println(e.Error())
	}
	_, e = c.New("round").Parse(matchesTemplate)
	if e != nil {
		fmt.Println(e.Error())
	}
	e = t.Execute(w, tournament)
	if e != nil {
		fmt.Println(e.Error())
	}
}

func recordResult(w http.ResponseWriter, r *http.Request) {
	roundNum, rErr := strconv.ParseInt(r.FormValue("round"), 10, 0)
	matchNum, mErr := strconv.ParseInt(r.FormValue("match"), 10, 0)

	if mErr != nil || rErr != nil {
		seeOther(w, "/matches")
	}

	match := &(tournament.Rounds[roundNum-1].Matches[matchNum-1])

	if r.Method == "POST" {

		result := r.FormValue("winner")
		var winner *Player
		if result == "corp" {
			winner = match.Game.Corp
		} else if result == "runner" {
			winner = match.Game.Runner
		}

		var timed bool
		if r.FormValue("timed") != "" {
			timed = true
		}

		match.Game.RecordResult(winner, timed)

		seeOther(w, "/matches")
	} else {
		data := map[string]string{"recordurl": r.URL.Path}
		data["roundNum"] = r.FormValue("round")
		data["matchNum"] = r.FormValue("match")
		data["corp"] = match.Corp.Name
		data["runner"] = match.Runner.Name

		if match.Game.Concluded {
			if match.Game.CorpWin {
				data["corpWin"] = "corpWin"
			} else if match.Game.RunnerWin {
				data["runnerWin"] = "runnerWin"
			} else {
				data["tie"] = "tie"
			}

			if match.Game.ModifiedWin {
				data["timed"] = "timed"
			}
		}
		e := applyTemplate(w, recordMatchTemplate, data)
		if e != nil {
			seeOther(w, "/matches")
		}
	}
}

func seeOther(w http.ResponseWriter, l string) {
	w.Header().Set("Location", l)
	w.WriteHeader(http.StatusSeeOther)
}

func main() {
	rand.Seed(time.Now().UnixNano())

	http.HandleFunc("/", menu)
	http.HandleFunc("/players", playerList)
	http.HandleFunc("/players/add", playerForm)
	http.HandleFunc("/players/change", changePlayer)
	http.HandleFunc("/standings", standings)
	http.HandleFunc("/matches", matches)
	http.HandleFunc("/rounds", rounds)
	http.HandleFunc("/recordResult", recordResult)
	http.HandleFunc("/finishRound", finishRound)
	http.HandleFunc("/nextRound", startRound)
	http.ListenAndServe("localhost:8080", nil)
}
