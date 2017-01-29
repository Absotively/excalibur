package main

import (
	"html/template"
	"net/http"
	"strconv"
)

var tournament Tournament

func playerList(w http.ResponseWriter, r *http.Request) {
	t, _ := template.New("players").Parse(playerListTemplate)
	t.Execute(w, tournament)
}

func standings(w http.ResponseWriter, r *http.Request) {
	t, _ := template.New("players").Parse(standingsTemplate)
	t.Execute(w, tournament)
}

func addPlayer(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{"saveurl": r.URL.Path}
	if r.Method == "POST" {
		name := r.FormValue("name")
		corp := r.FormValue("corp")
		runner := r.FormValue("runner")

		e := tournament.AddPlayer(name, corp, runner)

		if e == nil {
			seeOther(w, "/players")
			return
		}

		data["error"] = e.Error()
		data["name"] = name
		data["corp"] = corp
		data["runner"] = runner
	}

	t, _ := template.New("addPlayer").Parse(addPlayerTemplate)
	t.Execute(w, data)
}

func menu(w http.ResponseWriter, r *http.Request) {
	t, _ := template.New("menu").Parse(menuTemplate)
	t.Execute(w, nil)
}

func startRound(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		e := tournament.NextRound()
		if e != nil {
			t, _ := template.New("error").Parse(errorTemplate)
			t.Execute(w, e)
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
			t, _ := template.New("error").Parse(errorTemplate)
			t.Execute(w, e)
			return
		}
	}
	seeOther(w, "/")
}

func matches(w http.ResponseWriter, r *http.Request) {
	if len(tournament.Rounds) == 0 || len(tournament.Rounds[len(tournament.Rounds)-1].Matches) == 0 {
		t, _ := template.New("matches").Parse(noMatchesTemplate)
		t.Execute(w, nil)
	} else {
		t, _ := template.New("matches").Parse(matchesTemplate)
		t.Execute(w, tournament.Rounds[len(tournament.Rounds)-1])
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
		t, _ := template.New("result").Parse(recordMatchTemplate)
		t.Execute(w, data)
	}
}

func seeOther(w http.ResponseWriter, l string) {
	w.Header().Set("Location", l)
	w.WriteHeader(http.StatusSeeOther)
}

func main() {
	http.HandleFunc("/", menu)
	http.HandleFunc("/players", playerList)
	http.HandleFunc("/players/add", addPlayer)
	http.HandleFunc("/standings", standings)
	http.HandleFunc("/matches", matches)
	http.HandleFunc("/recordResult", recordResult)
	http.HandleFunc("/finishRound", finishRound)
	http.HandleFunc("/nextRound", startRound)
	http.ListenAndServe(":8080", nil)
}
