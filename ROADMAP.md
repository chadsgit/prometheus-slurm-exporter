# Roadmap

## v3: slurmrestd-based rewrite

Current architecture (`internal/collectors/`) shells out to `sinfo`, `squeue`,
`sdiag`, `sshare`, and `sacct`, then regex-parses the text output. This works
and is now well-tested, but it's fragile to Slurm version drift in CLI output
formatting, and requires the exporter to run on a node with the Slurm CLI
tools installed.

Slurm has shipped `slurmrestd` since 20.02, a REST API daemon that exposes
the same data (nodes, partitions, jobs, scheduler diagnostics, fairshare) as
versioned JSON endpoints (e.g. `/slurm/v0.0.40/...`) under an OpenAPI spec,
authenticated via JWT (`auth/jwt` plugin).

A v3 rewrite would:

- Replace `internal/slurm/exec.go` with an HTTP client against `slurmrestd`
- Replace each collector's regex parser with JSON unmarshalling into the
  OpenAPI-generated types
- Keep the exact same `slurm_*` metric names and labels, so
  `grafana/slurm-dashboard.json` and existing Prometheus alerting/recording
  rules keep working unchanged
- Add config for the `slurmrestd` URL and JWT token (env var or file)
