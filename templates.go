package main

const menuTemplate = `<h1>Tournament menu</h1>
<ul>
<li><a href="/players">Players</a></li>
<li><a href="/standings">Standings</a></li>
<li><a href="/matches">Current Round Matches</a></li>
<li><form action="/finishRound" method="POST"><input type="submit" value="Finish round"></li>
<li><form action="/nextRound" method="POST"><input type="submit" value="Start next round"></li>
</ul>
`

const playerListTemplate = `<h1>Players</h1>
{{if .Players}}<ul>
{{range .Players}}<li>{{.Name}} ({{.Corp}}, {{.Runner}})</li>
{{end}}
</table>
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
`

const addPlayerTemplate = `<h1>Add player</h1>
{{if .error}}<p><strong>Error: {{.error}}</p></strong>{{end}}
<form action="{{.saveurl}}" method="POST">
<label>Name: <input type="text" name="name" autofocus{{if .name}} value="{{.name}}"{{end}}></label><br>
<label>Corp: <input type="text" name="corp"{{if .corp}} value="{{.corp}}"{{end}}></label><br>
<label>Runner: <input type="text" name="runner"{{if .runner}} value="{{.runner}}"{{end}}></label><br>
<input type="submit" value="Add">
</form>
`

const matchesTemplate = `{{$roundNum := .Number}}<h1>Round {{$roundNum}}</h1>
<table><tr><th>#</th><th>Corp</th><th>Runner</th><th>Result</th></tr>
{{range .Matches}}
<tr>
<th>{{.Number}}</th>
<td{{if .Game.CorpWin}} class="winner"{{end}}>{{.Game.Pairing.Corp.Name}}</td>
<td{{if .Game.RunnerWin}} class="winner"{{end}}>{{.Game.Pairing.Runner.Name}}</td>
<td>{{if .Game.Concluded}}{{if or .Game.CorpWin .Game.RunnerWin}}{{if .Game.CorpWin}}Corp win{{else}}Runner win{{end}}{{if.Game.ModifiedWin}} (time){{end}}{{else}}Tie{{end}}{{else}}<a href="/recordResult?round={{$roundNum}}&match={{.Number}}">record</a>{{end}}</td>
</tr>
{{end}}
</table>
`

const noMatchesTemplate = `<h1>Matches</h1>
<p>No matches</p>
`

const recordMatchTemplate = `<h1>Record match result</h1>
<form action="{{.recordurl}}" method="POST">
<input type="hidden" name="round" value="{{.roundNum}}">
<input type="hidden" name="match" value="{{.matchNum}}">
<p>Winner:</p>
<label><input type="radio" name="winner" value="corp"> {{.corp}} (Corp)</label><br>
<label><input type="radio" name="winner" value="tie"> Tie</label><br>
<label><input type="radio" name="winner" value="runner"> {{.runner}} (Runner)</label></p>
<p><label><input type="checkbox" name="timed"> Timed/modified win</label></p>
<p><input type="submit" value="Record"></p>
</form>
`

const errorTemplate = `{{if .}}<p><strong>Error: {{.}}</strong></p>{{end}}`
