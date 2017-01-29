package main

const menuTemplate = `<h1>Tournament menu</h1>
<ul>
<li><a href="/players">Players</a></li>
<li><a href="/standings">Standings</a></li>
<li><a href="/matches">Current Round Matches</a></li>
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
<label>Name: <input type="text" name="name" autofocus{{if .name}} value="{{.name}}"{{end}}></label>
<label>Corp: <input type="text" name="corp"{{if .corp}} value="{{.corp}}"{{end}}></label>
<label>Runner: <input type="text" name="runner"{{if .runner}} value="{{.runner}}"{{end}}></label>
<input type="submit" value="Add">
</form>
`

const matchesTemplate = `<h1>Round {{.Number}}</h1>
<table><tr><th>Corp</th><th>Runner</th><th>Result</th></tr>
{{range .Matches}}
<tr>
<td{{if .Game.CorpWin}} class="winner"{{end}}>{{.Game.Pairing.Corp.Name}}</td>
<td{{if .Game.RunnerWin}} class="winner"{{end}}>{{.Game.Pairing.Runner.Name}}</td>
<td>{{if .Game.Concluded}}{{if .Game.CorpWin or .Game.RunnerWin}}{{if .Game.CorpWin}}Corp win{{else}}Runner win{{end}}{{if.Game.ModifiedWin}} (time){{end}}{{end}}Tie{{end}}</td>
</tr>
{{end}}
</table>
`

const noMatchesTemplate = `
<h1>Matches</h1>
<p>No matches</p>
`

const errorTemplate = `{{if .}}<p><strong>Error: {{.}}</strong></p>{{end}}`
