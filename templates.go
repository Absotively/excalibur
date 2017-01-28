package main

const playerListTemplate = `<h1>Players</h1>
{{if .Players}}<ul>
{{range .Players}}<li>{{.Name}} ({{.Corp}}, {{.Runner}})</li>
{{end}}
</ul>
{{end}}
<p><a href="/players/add">Add</a></p>
`

const addPlayerTemplate = `<h1>Add player</h1>
{{if .error}}<p><strong>Error: {{.error}}</p></strong>{{end}}
<form action="{{.saveurl}}" method="POST">
<label>Name: <input type="text" name="name"{{if .name}} value="{{.name}}"{{end}}></label>
<label>Corp: <input type="text" name="corp"{{if .corp}} value="{{.corp}}"{{end}}></label>
<label>Runner: <input type="text" name="runner"{{if .runner}} value="{{.runner}}"{{end}}></label>
<input type="hidden" name="process" value="true">
<input type="submit" value="Add">
</form>
`
