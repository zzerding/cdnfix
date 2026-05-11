# cdnfix

Tencent Cloud CDN refresh and push tool with explicit `site` and `job` management.

## Model

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
    file: ./urls/prod-a/refresh.txt

  - name: prod-b-push
    site: prod-b
    action: push
    file: ./urls/prod-b/push.txt
```

## Commands

Single site refresh:

```bash
cdnfix -e ./config/sites.yaml --site prod-a -f ./urls/prod-a/refresh.txt refresh
```

Single site push:

```bash
cdnfix -e ./config/sites.yaml --site prod-a -u https://example.com/a.js push
```

Batch jobs:

```bash
cdnfix -e ./config/sites.yaml -m ./config/jobs.yaml batch
```

Query pending tasks:

```bash
cdnfix -e ./config/sites.yaml query
```

Query one site only:

```bash
cdnfix -e ./config/sites.yaml --site prod-a query
```

## Runtime Files

Default runtime directories:

- logs: `./var/logs`
- task cache: `./var/cache`
- runs: `./var/runs`

Examples:

- log file: `var/logs/2026-05-11/prod-a.refresh.20260511T101530.log`
- task cache: `var/cache/prod-a/refresh.tasks.json`
- run record: `var/runs/2026-05-11/prod-a.refresh.20260511T101530.json`

Runtime directories can be overridden with:

- `--log-dir`
- `--cache-dir`
- `--run-dir`

## Development

Run tests:

```bash
go test ./...
```
