version: 1
sources:{{ range .Sources }}
  - channel:{{range .Channels }} {{.}} {{- end }}
    url: {{ .URL }}{{end}}
archs:{{ range .Archs }}
  - {{ . }}
{{end -}}
packages:{{ range .Packages }}
  - {{ . -}}
{{end}}

