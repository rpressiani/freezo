{{- if . }}
{{- range . }}
{{- if .Vulnerabilities }}
| Package | ID | Severity | Installed | Fixed | Title |
| --- | --- | --- | --- | --- | --- |
{{- range .Vulnerabilities }}
| `{{ .PkgName }}` | [{{ .VulnerabilityID }}]({{ .PrimaryURL }}) | {{ .Severity }} | {{ .InstalledVersion }} | {{ .FixedVersion }} | {{ .Title }} |
{{- end }}
{{- else }}
No vulnerabilities found.
{{- end }}
{{- end }}
{{- else }}
No vulnerabilities found.
{{- end }}
