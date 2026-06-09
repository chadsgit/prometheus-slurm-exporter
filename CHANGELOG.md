## Changelog

Full commit history per tag: https://github.com/vpenso/prometheus-slurm-exporter/commits/{tag number}

* **0.20** _(chadsgit fork. feat/harden-restructure)_
  - **Crash fix**: replaced all `log.Fatal`/`os.Exit` calls in collectors with `log.Printf` warnings. Exporter no longer dies on transient Slurm CLI errors (Bug #1)
  - **Crash fix**: bounds checks on `sinfo` per-node output prevent out-of-bounds panic on malformed lines (Bug #2)
  - **GPU crash fix**: bounds check on `sinfo` GPU output prevents panic when a node has no GRES (Bug #3)
  - **GPU crash fix**: NaN guard on GPU utilization calculation prevents invalid metric panic on zero-GPU clusters (Bug #4)
  - **GPU fix**: corrected `exec.Command` argument splitting for `sinfo` GPU query. Was being passed as a single string (Bug #5)
  - **GPU fix**: AllocTRES parser now handles both `gpu:N` (Slurm <23) and `gres/gpu=N` (Slurm 23+) formats (Bug #6)
  - **Scheduler fix**: `if/else if` guard on `Last cycle`/`Mean cycle` counters. Backfill values no longer mirror main scheduler values when backfilling is disabled (Bug #7)
  - **Restructure**: moved all collectors to `internal/collectors/`, shared exec helper to `internal/slurm/`, entry point to `cmd/exporter/main.go`
  - **Flags**: added `-fairshare-acct` flag to opt-in to sshare fair-share metrics (mirrors existing `-gpus-acct` pattern)
  - **Endpoint**: added `/health` HTTP endpoint for liveness checks
  - All package-level `regexp.MustCompile` vars (previously compiled inside scrape loops)
  - Full test coverage for all parse functions with regression tests for each crash bug

* **0.19**
  - Merge PR#50

* **0.18**
  - Add CPU/Memory info per node (see PR#47)

* **0.17**
  - Add fair share collector

* **0.16**
  - Export more data per account/partition, fix squeue for pending jobs
  - Merge PR#34

* **0.15**
  - CPU allocation status per partition

* **0.14**
  - add stats about jobs per account/per user

* **0.13**
  - Merge pull request #32 from pdtpartners/faster-node-metrics

* **0.12**
  - Merge pull request #30 from omnivector-solutions/add_snap_packaging

* **0.11**
  - Merge PR#29
  - Add more backfill stats (see PR#27)

* **0.10**
  - Scheduler: keep track of the DBD agent queue size

* **0.9**
  - README: update to fix build problem raised with issue #26

* **0.8**
  - Merge pull request #21 from cleargray/command-paths

* **0.7**
  - Update scheduler.go (fix issue #18)

* **0.6**
  - Merge pull request #13 from rug-cit-hpc/master

* **0.5**
  - [BUG]: count all job states (issue #9)

* **0.4**
  - Merge pull request #8 from MatMaul/pending-dep

* **0.3**
  - Fix issue #4

* **0.2**
  - Fix issue #3

* **0.1** 
  - Basic prototype
  - Merge PR#2
  - Add Grafana dashboard
