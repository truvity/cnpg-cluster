{{/*
Full name for the database resources.
Truncated to 63 chars for DNS-label compliance.
*/}}
{{- define "cnpg-database.fullname" -}}
{{- printf "%s-%s" .Values.clusterName .Values.databaseName | trunc 63 | trimSuffix "-" -}}
{{- end -}}
