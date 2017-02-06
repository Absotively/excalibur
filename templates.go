package main

const frameTemplate = `<!DOCTYPE html>
<html>
<head>
<title>Excalibur - Netrunner tournament</title>
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
<li><form action="/finishRound" method="POST"><input type="submit" value="Finish round"></form></li>
<li><form action="/nextRound" method="POST"><input type="submit" value="Start next round"></form></li>
</ul>
`

const playerListTemplate = `<h1>Players</h1>
{{if .Players}}<table>
{{range .Players}}<form action="/players/change" method="POST"><input type="hidden" name="name" value="{{.Name}}"><tr><td>{{.Name}}{{if or .Corp .Runner}} ({{.Corp}}{{if and .Corp .Runner}}, {{end}}{{.Runner}}){{end}}</td><td><a href="/players/change?name={{.Name}}">edit</a></td><td>{{if .Dropped}}Dropped <input type="submit" name="re-add" value="Re-add">{{else}}<input type="submit" name="drop" value="Drop">{{end}}</td></tr></form>
{{end}}</table>
{{end}}
<p><a href="/players/add">Add player</a></p>
<p><a href="/">Menu</a></p>
`

const standingsTemplate = `<h1>Standings</h1>
{{if .Players}}<table id="standings">
<tr><th>Player</th><th>Pts</th><th>SoS</th><th>XSoS</th></tr>
{{range .Players}}<tr><td>{{.Name}}</td><td>{{.Prestige}}</td><td>{{.SoS}}</td><td>{{.XSoS}}</td></tr>
{{end}}
</table>
{{end}}
<p><a href="/">Menu</a></p>
`

const playerFormTemplate = `<h1>{{if .add}}Add{{else}}Edit{{end}} player</h1>
{{if .error}}<p><strong>Error: {{.error}}</p></strong>{{end}}
<form action="{{.saveurl}}" method="POST">
<label>Name: <input type="text" name="name" autofocus{{if .name}} value="{{.name}}"{{end}}></label><br>
{{- if or .name .oldName}}<input type="hidden" name="old-name" value="{{if .oldName}}{{.oldName}}{{else}}{{.name}}{{end}}">{{end -}}
<label>Corp: <input type="text" name="corp"{{if .corp}} value="{{.corp}}"{{end}}></label><br>
<label>Runner: <input type="text" name="runner"{{if .runner}} value="{{.runner}}"{{end}}></label><br>
<input type="submit" {{if .add}}name="add" value="Add"{{else}}name="edit" value="Change"{{end}}>
</form>
`

const matchesTemplate = `{{$roundNum := .Number}}<h1>Round {{$roundNum}}</h1>
<table><tr><th>#</th><th>Corp</th><th>Runner</th><th>Result</th></tr>
{{range .Matches}}
<tr>
<th>{{.Number}}</th>
<td class="corp
 {{- if .Game.CorpWin}} winner{{end -}}
">{{.Game.Pairing.Corp.Name}}</td>
<td class="runner
 {{- if .Game.RunnerWin}} winner{{end -}}
 {{- if not .Game.Pairing.Runner}} bye{{end -}}
">
 {{- if .Game.Pairing.Runner -}}
  {{.Game.Pairing.Runner.Name}}
 {{- else -}}
  BYE
 {{- end -}}
</td>
<td class="result
 {{- if not .Game.Pairing.Runner}} bye{{end -}}
">
 {{- if not .Game.Pairing.Runner -}}
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
 {{if .Game.Pairing.Runner}}
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
