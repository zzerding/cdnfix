# cdnfix

Tencent Cloud CDN refresh and push tool with explicit `site` and `job` management.

## Quick Start

Get command help:

```bash
cdnfix --help
cdnfix refresh --help
cdnfix push --help
cdnfix batch --help
cdnfix query --help
```

Typical commands:

```bash
cdnfix --root /opt/cdnfix batch
cdnfix --root /opt/cdnfix --site prod-a -f urls/prod-a/refresh.txt refresh
cdnfix --root /opt/cdnfix --site prod-a -u https://example.com/a.js push
cdnfix --root /opt/cdnfix query
```

## Model

Configuration is anchored to a single application root. By default, the root is the directory that contains the `cdnfix` executable, so config and runtime files do not depend on the current working directory.

Configuration is split into three parts:

- `sites`: credentials and region per site
- `jobs`: mapping from `site` to URL file and action
- `var/`: runtime output for logs, task cache, and run metadata

Recommended layout:

```text
config/
  sites.yaml
  jobs.yaml

urls/
  prod-a/
    refresh.txt
    push.txt
  prod-b/
    refresh.txt

var/
  logs/
  cache/
  runs/
```

If the binary is not deployed inside the root directory, pass `--root /path/to/app`.

### use source run commad
1. `git clone https://github.com/zzerding/cdnfix.git`
2. cd cdnfix
3. go run main.go

### use binary run commad
1. go install  github.com/zzerding/cdnfix
2. cdnfix 

### use docker run commad
1.use -e args
```
 docker run -rm -v $PWD:/app -e SECRET_ID=xxx -e SECRET_KEY=xxx zzerding/refresh-cnd 
```
2. save cache to local
```
 docker run --rm --env-file=.env -v $(PWD)/.task_push.cache:/root/.task_push.cache -v $(PWD)/.task_refresh.cache:/root/.task_refresh.cache zzerding/cdnfix -u https://www.xxxx.com/join/ push
```
3. query status
```
 docker run --rm  --env-file=.env  -v $(PWD)/.task_push.cache:/root/.task_push.cache -v $(PWD)/.task_refresh.cache:/root/.task_refresh.cache zzerding/cdnfix  query
```

## Site Config

`config/sites.yaml`

```yaml
sites:
  prod-a:
    secret_id: your-secret-id
    secret_key: your-secret-key
    region: ap-guangzhou

  prod-b:
    secret_id: your-secret-id
    secret_key: your-secret-key
    region: ap-shanghai
```

Single-site `.env` is still supported:

```env
SECRET_ID=your-secret-id
SECRET_KEY=your-secret-key
REGION=ap-guangzhou
```

## Jobs Config

`config/jobs.yaml`

```yaml
jobs:
  - name: prod-a-refresh
    site: prod-a
    action: refresh
    file: ../urls/prod-a/refresh.txt

  - name: prod-b-push
    site: prod-b
    action: push
    file: ../urls/prod-b/push.txt
```

`file` is resolved relative to `jobs.yaml`, not relative to the shell working directory.

## Commands

Single site refresh:

```bash
/opt/cdnfix/cdnfix --site prod-a -f urls/prod-a/refresh.txt refresh
```

Single site push:

```bash
/opt/cdnfix/cdnfix --site prod-a -u https://example.com/a.js push
```

Batch jobs:

```bash
/opt/cdnfix/cdnfix batch
```

Query pending tasks:

```bash
/opt/cdnfix/cdnfix query
```

Query one site only:

```bash
/opt/cdnfix/cdnfix --site prod-a query
```

Override the root explicitly when needed:

```bash
cdnfix --root /opt/cdnfix batch
```

Command help is also available from the CLI:

```bash
cdnfix --help
cdnfix refresh --help
cdnfix push --help
cdnfix batch --help
cdnfix query --help
```

## Runtime Files

Default runtime directories:

- logs: `<root>/var/logs`
- task cache: `<root>/var/cache`
- runs: `<root>/var/runs`

Examples:

- log file: `/opt/cdnfix/var/logs/2026-05-11/prod-a.refresh.20260511T101530.log`
- task cache: `/opt/cdnfix/var/cache/prod-a/refresh.tasks.json`
- run record: `/opt/cdnfix/var/runs/2026-05-11/prod-a.refresh.20260511T101530.json`

Runtime directories can be overridden with:

- `--log-dir`
- `--cache-dir`
- `--run-dir`

Configuration path defaults:

- site config: `<root>/config/sites.yaml`
- fallback single-site env: `<root>/config/.env`, then `<root>/.env`
- jobs manifest: `<root>/config/jobs.yaml`

## Development

Run tests:

```bash
go test ./...
```
