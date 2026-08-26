{{/*
Resolve the serverName for recovery source.
Defaults to clusterName if not explicitly set.
*/}}
{{- define "cnpg-cluster.sourceServerName" -}}
{{- /* No dig: .Values is chartutil.Values, which dig rejects
(interface conversion panic — found in the first live recovery
drill). Chained defaults survive absent intermediate keys. */ -}}
{{- $src := ((.Values.bootstrap | default dict).recovery | default dict).source | default dict -}}
{{- $src.serverName | default .Values.clusterName -}}
{{- end -}}

{{/*
The hba posture (gitops docs/architecture/pg-usage-shapes.md): hostssl
only, never trust, implicit reject at the end. hba is FIRST-MATCH, so
password (scram) lines are emitted per password role BEFORE the
catch-all cert line — a leading blanket cert rule would shadow scram and
break every password login. Projects may APPEND via
postgresql.extra_pg_hba, never replace. The waist OAuth line lands with
the validator (INF-479).
*/}}
{{- define "cnpg-cluster.pgHba" -}}
{{- if .Values.bootstrap.initdb.owner }}
- hostssl all {{ .Values.bootstrap.initdb.owner }} all scram-sha-256
{{- end }}
{{- range .Values.roles }}
{{- if and (eq (.auth | default "cert") "password") (ne .name $.Values.bootstrap.initdb.owner) }}
- hostssl all {{ .name | replace "-" "_" }} all scram-sha-256
{{- end }}
{{- end }}
- hostssl all all all cert
{{- range .Values.postgresql.extra_pg_hba }}
- {{ . }}
{{- end }}
{{- end -}}

{{/*
Default s3 prefix = {namespace}/{clusterName}/ — the backup permission
model's write scope.
*/}}
{{- define "cnpg-cluster.s3Prefix" -}}
{{- .Values.backup.s3Prefix | default (printf "%s/%s" .Values.namespace .Values.clusterName) -}}
{{- end -}}
