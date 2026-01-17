{{- if . }}
{{- range . }}
{{- if .Vulnerabilities }}
{{- range .Vulnerabilities }}
#### {{ .VulnerabilityID }} - {{ .Title }}
**Severity:** {{ .Severity }}
**Package:** `{{ .PkgName }}`
**Installed Version:** `{{ .InstalledVersion }}`
**Fixed Version:** `{{ .FixedVersion }}`

> {{ .Description }}

---
{{- end }}
{{- end }}
{{- end }}
{{- end }}
