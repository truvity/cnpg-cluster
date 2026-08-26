# cnpg-cluster

Helm charts for running PostgreSQL on [CloudNativePG](https://cloudnative-pg.io/):

- **charts/cnpg-cluster** — a CNPG `Cluster` with an opinionated two-posture
  profile model, plus roles, declared databases, backup `ObjectStore`s and
  `ScheduledBackup`s.
- **charts/cnpg-database** — a logical `Database` inside an existing cluster
  (CNPG 1.28+ declarative database management), for consumers that own their
  database but not the cluster.

Published as OCI charts:

```
oci://ghcr.io/truvity/charts/cnpg-cluster
oci://ghcr.io/truvity/charts/cnpg-database
```

## The profile contract

`profile` is **required** and must be `devel` or `prod` — a missing value fails
the render rather than silently choosing a posture:

- `devel` — 1 instance, `enablePDB: false`. A single-instance cluster has
  nowhere to fail over; a PDB would only stall node maintenance.
- `prod` — 3 instances, `enablePDB: true`. Replicas make eviction survivable
  via switchover; the PDB protects the primary while Karpenter (or any drainer)
  moves replicas freely.

`enablePDB` is always emitted explicitly (CNPG's server-side default is `true`,
so omission on devel would resurrect the drain-blocking primary PDB). The chart
tests in `charts/cnpg-cluster/chart_test.go` pin this.

## Scheduling

`scheduling.databasePool` (default `database`) selects the Karpenter node pool
CNPG instances land on and tolerates a taint of the same name. Set it to your
pool name (e.g. `durable` for pools named by interruption contract), or to the
empty string to opt out entirely on clusters without a dedicated database pool.

## Development

```sh
devbox shell   # or rely on direnv
just check     # test + lint + chart-lint + vuln
```

Chart versions are placeholders (`0.0.0`); the release workflow injects the
real version from the git tag via `ocictl helmctl` when packaging.

## Provenance

Extracted from Truvity's internal monorepo (the `nexus` charts) so the cluster
and database shapes can be consumed by any estate, MIT-licensed.

## License

MIT — see [LICENSE](LICENSE).
