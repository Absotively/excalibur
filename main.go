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
	applyTemplate(w, standingsTemplate, &tournament)
}

func playerForm(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{"saveurl": r.URL.Path}
	edit := (r.FormValue("edit") != "" || (r.FormValue("player-id") != "" && r.Method == "GET"))
	if !edit {
		data["add"] = "add"
	}

	var e error
	id := NoPlayer
	name := r.FormValue("name")
	corp := r.FormValue("corp")
	runner := r.FormValue("runner")
	idString := r.FormValue("player-id")
	if idString != "" {
		idTemp, err := strconv.Atoi(idString)
		if err == nil {
			id = PlayerID(idTemp)
		}
	}

	if r.Method == "POST" {
		if edit {
			player := tournament.Player(id)
			if player != nil {
				player.Name = name
				player.Corp = corp
				player.Runner = runner
			} else {
				e = errors.New("No such player")
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
	if edit && r.Method == "GET" {
		player := tournament.Player(id)
		name = player.Name
		corp = player.Corp
		runner = player.Runner
	}

	if e != nil {
		data["error"] = e.Error()
	}
	data["name"] = name
	data["corp"] = corp
	data["runner"] = runner
	data["id"] = idString
	if !edit {
		data["add"] = "add"
	}

	applyTemplate(w, playerFormTemplate, data)
}

func changePlayer(w http.ResponseWriter, r *http.Request) {
	idString := r.FormValue("player-id")

	if idString != "" && (r.Method == "GET" || r.FormValue("edit") != "") {
		playerForm(w, r)
		return
	}

	if r.Method != "POST" {
		seeOther(w, "/players")
		return
	}

	id := NoPlayer
	if idString != "" {
		idTemp, err := strconv.Atoi(idString)
		if err == nil {
			id = PlayerID(idTemp)
		}
	}

	if r.FormValue("drop") != "" {
		tournament.DropPlayer(id)
	} else if r.FormValue("re-add") != "" {
		tournament.ReAddPlayer(id)
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
		var winner PlayerID
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
		data["corp"] = tournament.Player(match.Corp).Name
		data["runner"] = tournament.Player(match.Runner).Name

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
