# cdnfix

Tencent Cloud CDN refresh and push tool with explicit `site` and `job` management.

## Quick Start

Get command help:

```bash
cdnfix -v
cdnfix --help
cdnfix init --help
cdnfix refresh --help
cdnfix push --help
cdnfix batch --help
cdnfix query --help
```

## System Install Layout

For a general system installation, use fixed config, state, and log directories instead of relying on the current working directory.

Recommended layout:

```text
/etc/cdnfix/
  sites.yaml
  jobs.yaml
  urls/
    prod-a/
      refresh.txt
      push.txt
    prod-b/
      refresh.txt

/var/lib/cdnfix/
  cache/
  runs/

/var/log/cdnfix/
```

Default config files:

- `/etc/cdnfix/sites.yaml`
- `/etc/cdnfix/jobs.yaml`

Runtime output:

- task cache: `/var/lib/cdnfix/cache`
- run metadata: `/var/lib/cdnfix/runs`
- logs: `/var/log/cdnfix`

Typical commands:

```bash
cdnfix init
cdnfix batch
cdnfix batch --job prod-a-refresh
cdnfix --site prod-a -f /etc/cdnfix/urls/prod-a/refresh.txt refresh
cdnfix --site prod-a -u https://example.com/a.js push
printf '%s\n' https://example.com/a.js | cdnfix --site prod-a push
cdnfix query
cdnfix --site prod-a query
```

## URL Input Methods

`refresh` and `push` support three URL input methods:

1. `--urls` / `-u`: comma-separated URLs
2. `--urlfile` / `-f`: one URL per line from a file
3. `stdin`: one URL per line from a shell pipeline

When stdin is piped in, `cdnfix` reads from stdin first.

Examples:

```bash
# Comma-separated URLs
cdnfix --site prod-a -u https://example.com/a.js,https://example.com/b.css push

# Read from file
cdnfix --site prod-a -f /etc/cdnfix/urls/prod-a/refresh.txt refresh

# Read from stdin
printf '%s\n' https://example.com/a.js https://example.com/b.css | cdnfix --site prod-a push
cat /etc/cdnfix/urls/prod-a/refresh.txt | cdnfix --site prod-a refresh
```

Initialize a fresh layout:

```bash
cdnfix init
cdnfix --root /opt/cdnfix init
cdnfix --root /opt/cdnfix --site-name prod-a init
```

## Path Precedence

Path resolution should follow this order:

1. CLI flags
   - `--config-dir`
   - `--state-dir`
   - `--log-dir`
   - `--root`
2. Environment variables
   - `CDNFIX_CONFIG_DIR`
   - `CDNFIX_STATE_DIR`
   - `CDNFIX_LOG_DIR`
   - `CDNFIX_ROOT`
3. Portable root shortcut
   - `--root` or `CDNFIX_ROOT` maps to a self-contained layout under one directory
4. System defaults
   - config: `/etc/cdnfix`
   - state: `/var/lib/cdnfix`
   - logs: `/var/log/cdnfix`

Use explicit CLI flags or environment variables when the installation does not follow the system defaults.

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

System install example: `/etc/cdnfix/sites.yaml`

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

Single-site `.env` fallback is still supported when you only need one site.

## Jobs Config

System install example: `/etc/cdnfix/jobs.yaml`

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

`file` paths are resolved relative to the manifest file, not relative to the shell working directory.

That means `file: ./urls/prod-a/refresh.txt` in `/etc/cdnfix/jobs.yaml` resolves to `/etc/cdnfix/urls/prod-a/refresh.txt`.

## Portable Deployment

Portable deployment is still supported. In that mode, keep everything under one directory and pass `--root` explicitly.

Example layout:

```text
/opt/cdnfix/
  cdnfix
  config/
    sites.yaml
    jobs.yaml
    urls/
      prod-a/
        refresh.txt
  var/
    lib/
      cache/
      runs/
    log/
```

Portable examples:

```bash
cdnfix --root /opt/cdnfix batch
cdnfix --root /opt/cdnfix --site prod-a -f /opt/cdnfix/config/urls/prod-a/refresh.txt refresh
cdnfix --root /opt/cdnfix query
```

With `--root /opt/cdnfix`, the portable layout is:

- site config: `/opt/cdnfix/config/sites.yaml`
- jobs manifest: `/opt/cdnfix/config/jobs.yaml`
- task cache: `/opt/cdnfix/var/lib/cache`
- run metadata: `/opt/cdnfix/var/lib/runs`
- logs: `/opt/cdnfix/var/log`

## Development

Run tests:

```bash
go test ./...
```
