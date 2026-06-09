# Development

Setup the development environment on a node with access to the Slurm user
command-line interface, in particular with the `sinfo`, `squeue`, and `sdiag`
commands.

## Install Go from source

```bash
export VERSION=1.15 OS=linux ARCH=amd64
wget https://dl.google.com/go/go$VERSION.$OS-$ARCH.tar.gz
tar -xzvf go$VERSION.$OS-$ARCH.tar.gz
export PATH=$PWD/go/bin:$PATH
```

_Alternatively install Go using the packaging system of your Linux distribution._

## Clone this repository and build

```bash
git clone -b feat/harden-restructure https://github.com/chadsgit/prometheus-slurm-exporter
cd prometheus-slurm-exporter
go build ./cmd/exporter/
```

To run the tests:

```bash
go test ./...
```

Start the exporter (foreground) and query all metrics:

```bash
./exporter
curl http://localhost:8080/metrics
curl http://localhost:8080/health
```

Optional flags:

```bash
./exporter --listen-address="0.0.0.0:<port>"   # change listen port (default 8080)
./exporter --gpus-acct                          # enable GPU accounting metrics
./exporter --fairshare-acct                     # enable fair-share metrics (requires sshare)
```

## References

* [GOlang Package Documentation](https://godoc.org/github.com/prometheus/client_golang/prometheus)
* [Metric Types](https://prometheus.io/docs/concepts/metric_types/)
* [Writing Exporters](https://prometheus.io/docs/instrumenting/writing_exporters/)
* [Available Exporters](https://prometheus.io/docs/instrumenting/exporters/)
