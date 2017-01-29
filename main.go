package main

import (
	"html/template"
	"net/http"
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
			w.Header().Set("Location", "/players")
			w.WriteHeader(http.StatusSeeOther)
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
			w.Header().Set("Location", "/matches")
			w.WriteHeader(http.StatusSeeOther)
		}
	} else {
		w.Header().Set("Location", "/")
		w.WriteHeader(http.StatusSeeOther)
	}
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

func main() {
	http.HandleFunc("/", menu)
	http.HandleFunc("/players", playerList)
	http.HandleFunc("/players/add", addPlayer)
	http.HandleFunc("/standings", standings)
	http.HandleFunc("/matches", matches)
	http.HandleFunc("/nextRound", startRound)
	http.ListenAndServe(":8080", nil)
}
