# kubectl-refx

`kubectl-refx` is a `kubectl` plugin that helps you answer:

- Which workloads reference this ConfigMap or Secret?
- Which ConfigMaps or Secrets are unused?
- How many workloads reference each resource?

It scans Deployments, StatefulSets, DaemonSets, Jobs, CronJobs, and Pods, and detects references from:

- `env` (`valueFrom.configMapKeyRef` / `valueFrom.secretKeyRef`)
- `envFrom` (`configMapRef` / `secretRef`)
- `volumes` (`configMap` / `secret`)
- Both containers and init containers

## Features

- Supports `ConfigMap`, `Secret`, auto-detect mode (`detect`), aggregate listing (`list`), and orphan discovery (`unused`)
- Namespace-scoped or cluster-wide scans (`-n`, `-A`)
- Multiple resource names in one call (`configmap <name> [name...]`, `secret <name> [name...]`, `detect <name> [name...]`)
- Optional workload filtering by kind (`-K`) and label selector (`-l`)
- Watch mode (`--watch --interval`) for repeated scans
- Output formats: `table`, `wide`, `json`, `yaml`
- Optional header suppression for table output (`--no-headers`)
- Excludes common auto-generated system resources by default (with opt-in `--include-system`)
- Correctly handles same-name resources across namespaces (for example reflector-style copies)

## Requirements

- Go `1.25+` (for building from source)
- `kubectl` configured with cluster access (`~/.kube/config` or `$KUBECONFIG`)

## Install

### Build from source

```bash
git clone https://github.com/tlmanz/kubectl-refx.git
cd kubectl-refx
make build
```

Binary output:

- `./bin/kubectl-refx`

### Install into PATH for kubectl plugin discovery

```bash
make build
sudo make install
```

This installs to:

- `/usr/local/bin/kubectl-refx`

### Verify plugin discovery

```bash
kubectl plugin list | grep refx
kubectl refx --help
```

## Command Overview

```bash
kubectl refx configmap <name> [name...]
kubectl refx secret <name> [name...]
kubectl refx detect <name> [name...]
kubectl refx list configmap
kubectl refx list secret
kubectl refx unused configmap
kubectl refx unused secret
```

Aliases:

- `configmap`: `cm`, `configmaps`
- `secret`: `sec`, `secrets`
- `detect`: `auto`, `find`

## Global Flags

| Flag | Short | Default | Description |
|---|---|---|---|
| `--namespace` | `-n` | current kube context namespace | Target namespace |
| `--all-namespaces` | `-A` | `false` | Scan across all namespaces |
| `--output` | `-o` | `table` | `table`, `wide`, `json`, `yaml` |
| `--timeout` |  | `30s` | Request timeout per scan |
| `--context` |  | current kube context | Override kubeconfig context |
| `--kind` | `-K` | all supported kinds | Comma-separated kinds (e.g. `Deployment,Job`) |
| `--selector` | `-l` | empty | Label selector for workloads |
| `--no-headers` |  | `false` | Hide table header row |
| `--watch` |  | `false` | Re-run scan on an interval |
| `--interval` |  | `5s` | Watch interval |
| `--include-system` |  | `false` | Include system-generated ConfigMaps/Secrets |

Supported workload kinds for `-K`:

- `Deployment`
- `StatefulSet`
- `DaemonSet`
- `Job`
- `CronJob`
- `Pod`

## Usage Examples

### Find references to one ConfigMap

```bash
kubectl refx configmap shared-mos-acl-provider-default -n mos
```

### Find references to multiple Secrets

```bash
kubectl refx secret db-credentials redis-credentials -n production
```

### Auto-detect whether each resource is a ConfigMap or Secret

```bash
kubectl refx detect app-config db-password -n staging
```

### Scan all namespaces

```bash
kubectl refx configmap shared-mos-acl-provider-default -A
```

### Show container names in output (`wide`)

```bash
kubectl refx secret api-key -n platform -o wide
```

### Filter to selected workload kinds only

```bash
kubectl refx configmap app-config -n prod -K Deployment,StatefulSet
```

### Filter workloads by label selector

```bash
kubectl refx secret db-password -A -l 'app in (api,worker)'
```

### Watch for changes every 10 seconds

```bash
kubectl refx configmap app-config -n prod --watch --interval 10s
```

### List all ConfigMaps with reference counts

```bash
kubectl refx list configmap -n prod
```

### List all Secrets with reference counts across the cluster

```bash
kubectl refx list secret -A
```

### Find unused Secrets

```bash
kubectl refx unused secret -n prod
```

### Machine-readable output

```bash
kubectl refx list configmap -A -o json
kubectl refx unused secret -n prod -o yaml
```

## Output Formats

### `table` (default)

Columns for `configmap` / `secret` / `detect`:

- `NAMESPACE`
- `KIND`
- `NAME`
- `REFERENCE TYPE`
- `DETAIL`

### `wide`

Same as `table`, plus:

- `CONTAINER`

### `json` / `yaml`

- Structured output for scripting and pipelines
- `list` and `unused` return per-resource rows (`namespace`, `name`, `references` or just `namespace`, `name`)

## Notes on Reflected or Same-Name Resources

If you use reflector tools and the same resource name exists in multiple namespaces, `kubectl-refx` treats each resource as namespace-qualified. That means:

- `list` and `unused` keep counts per actual namespace
- Same-name resources are not merged into one row

## System Resources Behavior

By default, `kubectl-refx` excludes common auto-generated resources:

- ConfigMap: `kube-root-ca.crt`
- Secret types:
  - `kubernetes.io/service-account-token`
  - `kubernetes.io/dockercfg`
  - `kubernetes.io/dockerconfigjson`

Use `--include-system` to include them.

## Development

```bash
make build
make test
make lint
make clean
```

## Dependencies

Primary dependencies:

- `k8s.io/client-go`
- `k8s.io/api`
- `k8s.io/apimachinery`
- `github.com/spf13/cobra`
- `github.com/olekukonko/tablewriter`
- `golang.org/x/sync`
- `sigs.k8s.io/yaml`

## License

Apache 2.0. See [LICENSE](LICENSE).
