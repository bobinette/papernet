package mail

const mailTemplate = `
{{ define "link" }}<a href="{{.}}">{{.}}</a>{{end}}
<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Transitional//EN"
        "http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd">
<html>
    </head>
    <body>
        <p>Your search "{{.Q}}" (on {{.Sources}}) has new results:</p>
        <ul>
            {{ range .Papers }}
                <li>{{ .Title }} - {{ template "link" index .References 0 }}</li>
            {{ end }}
        </ul>
        <p>Check them out at {{template "link" .Link}}.</p>
    </body>
</html>
`
