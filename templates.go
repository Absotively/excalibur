package main

const frameTemplate = `<!DOCTYPE html>
<html>
<head>
<title>Excalibur - Netrunner tournament</title>
<style>
table { border-collapse: collapse; }
td, th { padding: 0.4em 0.8em; border-bottom: 1px solid #aaaaaa; }
th { font-weight: bold }
td.winner {font-weight: bold }
td.corp { border-bottom: 2px solid #0000aa; }
td.runner { border-bottom: 2px solid #aa0000; }
li { padding-bottom: 0.4em; }
</style>
</head>
<body>
{{template "content" .}}
</body>
</html>
`

const menuTemplate = `<h1>Tournament menu</h1>
<ul>
<li><a href="/players">Players</a></li>
<li><a href="/standings">Standings</a></li>
<li><a href="/matches">Current Round Matches</a></li>
<li><a href="/rounds">All rounds</a></li>
<li><form action="/finishRound" method="POST"><input type="submit" value="Finish round"></form></li>
<li><form action="/nextRound" method="POST"><input type="submit" value="Start next round"></form></li>
</ul>
`

const playerListTemplate = `<h1>Players</h1>
{{if .Players}}<table>
{{range .Players}}<form action="/players/change" method="POST"><input type="hidden" name="player-id" value="{{.PlayerID}}"><tr><td>{{.Name}}{{if or .Corp .Runner}} ({{.Corp}}{{if and .Corp .Runner}}, {{end}}{{.Runner}}){{end}}</td><td><a href="/players/change?player-id={{.PlayerID}}">edit</a></td><td>{{if .Dropped}}Dropped <input type="submit" name="re-add" value="Re-add">{{else}}<input type="submit" name="drop" value="Drop">{{end}}</td></tr></form>
{{end}}</table>
{{end}}
<p><a href="/players/add">Add player</a></p>
<p><a href="/">Menu</a></p>
`

const standingsTemplate = `{{$t := .}}<h1>Standings</h1>
{{if .Standings}}<table id="standings">
<tr><th>Player</th><th>Pts</th><th>SoS</th><th>XSoS</th></tr>
{{range .Standings}}{{$p := ($t.Player .)}}<tr><td>{{$p.Name}}</td><td>{{$p.Prestige}}</td><td>{{printf "%.3f" $p.SoS}}</td><td>{{printf "%.3f" $p.XSoS}}</td></tr>
{{end}}
</table>
{{end}}
<p><a href="/">Menu</a></p>
`

const playerFormTemplate = `<h1>{{if .add}}Add{{else}}Edit{{end}} player</h1>
{{if .error}}<p><strong>Error: {{.error}}</p></strong>{{end}}
<form action="{{.saveurl}}" method="POST">
<label>Name: <input type="text" name="name" autofocus{{if .name}} value="{{.name}}"{{end}}></label><br>
{{- if .id}}<input type="hidden" name="player-id" value="{{.id}}">{{end -}}
<label>Corp: <input type="text" name="corp"{{if .corp}} value="{{.corp}}"{{end}}></label><br>
<label>Runner: <input type="text" name="runner"{{if .runner}} value="{{.runner}}"{{end}}></label><br>
<input type="submit" {{if .add}}name="add" value="Add"{{else}}name="edit" value="Change"{{end}}>
</form>
`

const matchesTemplate = `{{$t := .Tournament}}{{$roundNum := .Number}}<h1>Round {{$roundNum}}</h1>
<table><tr><th>#</th><th>Corp</th><th>Runner</th><th>Result</th></tr>
{{range .Matches}}
<tr>
<th>{{.Number}}</th>
<td class="corp
 {{- if .Game.CorpWin}} winner{{end -}}
">{{($t.Player .Game.Pairing.Corp).Name}}</td>
<td class="runner
 {{- if .Game.RunnerWin}} winner{{end -}}
 {{- if .IsBye}} bye{{end -}}
">
 {{- if not .IsBye -}}
  {{($t.Player .Game.Pairing.Runner).Name}}
 {{- else -}}
  BYE
 {{- end -}}
</td>
<td class="result
 {{- if not .IsBye}} bye{{end -}}
">
 {{- if .IsBye -}}
  BYE
 {{- else if .Game.Concluded}}
  {{- if or .Game.CorpWin .Game.RunnerWin}}
   {{- if .Game.CorpWin -}}
    Corp win
   {{- else -}}
    Runner win
   {{- end}}
   {{- if.Game.ModifiedWin}} (time)
   {{- end}}
  {{- else -}}
   Tie
  {{- end -}}
 {{- end -}}
 {{if not .IsBye}}
  {{- if .Game.Concluded}} ({{end -}}
   <a href="/recordResult?round={{$roundNum}}&match={{.Number}}">
   {{- if .Game.Concluded}}edit{{else}}record{{end -}}
   </a>
   {{- if .Game.Concluded}}){{end}}
  {{- end -}}
</td>
</tr>
{{end}}
</table>
<p><a href="/">Menu</a></p>
`

const noMatchesTemplate = `<h1>Matches</h1>
<p>No matches</p>
`

const roundsTemplate = `{{range .Rounds}}{{template "round" .}}{{end}}`

const recordMatchTemplate = `<h1>{{if .winner}}Update{{else}}Record{{end}} match result</h1>
<form action="{{.recordurl}}" method="POST">
<input type="hidden" name="round" value="{{.roundNum}}">
<input type="hidden" name="match" value="{{.matchNum}}">
<p>Winner:</p>
<label><input type="radio" name="winner" value="corp"{{if .corpWin}} checked{{end}}> {{.corp}} (Corp)</label><br>
<label><input type="radio" name="winner" value="tie"{{if .tie}} checked{{end}}> Tie</label><br>
<label><input type="radio" name="winner" value="runner"{{if .runnerWin}} checked{{end}}> {{.runner}} (Runner)</label></p>
<p><label><input type="checkbox" name="timed"{{if .timed}} checked{{end}}> Timed/modified win</label></p>
<p><input type="submit" value="Record"></p>
</form>
`

const errorTemplate = `{{if .}}<p><strong>Error: {{.}}</strong></p>{{end}}`
