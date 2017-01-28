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

func addPlayer(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{"saveurl": r.URL.Path}
	if r.FormValue("process") != "" {
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

func main() {
	http.HandleFunc("/players", playerList)
	http.HandleFunc("/players/add", addPlayer)
	http.ListenAndServe(":8080", nil)
}
