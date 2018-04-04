# katafygio

[![Build Status](https://travis-ci.org/bpineau/katafygio.svg?branch=master)](https://travis-ci.org/bpineau/katafygio)
[![Go Report Card](https://goreportcard.com/badge/github.com/bpineau/katafygio)](https://goreportcard.com/report/github.com/bpineau/katafygio)

**katafygio** discovers Kubernetes objects (deployments, services, ...),
and continuously save them as yaml files in a git repository.
This provides real time, continuous backups, and keeps detailled changes history.

## Usage

```bash
katafygio --git-url https://user:token@github.com/myorg/myrepos.git --local-dir /tmp/kfdump
```

The remote git url is optional: if not specified, katafygio will create a local
repository in the target dump directory. katafygio may be used that way to
quickly dump a cluster's content.

Filtering out irrelevant objects (esp. ReplicaSets and Pods) with `-x` and `-y`
will help to keep memory usage low. Same for low value and frequently changing
objects, that may bloat the git history.


```bash
# Filtering out replicasets and pods since they are managed by Deployments
# (already archived), endpoints (managed by Services), secrets (to keep them
# confidential), events and node (irrelevant), and the leader-elector
# configmap that has low value and changes a lot, causing commits churn.

katafygio \
  -g https://user:token@github.com/myorg/myrepos.git -e /tmp/kfdump \
  -x secret -x pod -x replicaset -x node -x endpoints -x event \
  -y configmap:kube-system/leader-elector
```

You can also use the [docker image](https://hub.docker.com/r/bpineau/katafygio/).

## CLI options

```
Flags:
  -s, --api-server string            Kubernetes api-server url
  -c, --config string                Configuration file (default "/etc/katafygio/katafygio.yaml")
  -d, --dry-run                      Dry-run mode: don't store anything.
  -x, --exclude-kind stringSlice     Ressource kind to exclude. Eg. 'deployment' (can be specified several times)
  -y, --exclude-object stringSlice   Object to exclude. Eg. 'configmap:kube-system/kube-dns' (can be specified several times)
  -l, --filter string                Label filter. Select only objects matching the label.
  -g, --git-url string               Git repository URL
  -p, --healthcheck-port int         Port for answering healthchecks on /health
  -h, --help                         help for katafygio
  -k, --kube-config string           Kubernetes config path
  -e, --local-dir string             Where to dump yaml files (default "./kubernetes-backup")
  -v, --log-level string             Log level (default "info")
  -o, --log-output string            Log output (default "stderr")
  -r, --log-server string            Log server (if using syslog)
  -i, --resync-interval int          Resync interval in seconds (0 to disable) (default 300)
```

## Config file and env variables

All settings can be passed by cli, or environment variable, or in a yaml configuration file
(thanks to Viper and Cobra libs). The environment are the same as cli options, in uppercase,
prefixed by "KF", and with underscore instead of dashs. ie.:

```
export KF_GIT_URL=https://user:token@github.com/myorg/myrepos.git
export KF_LOCAL_DIR=/tmp/kfdump
```

## Build

Assuming you have go 1.10 and glide in the path, and GOPATH configured:

```shell
make deps
make build
```

## See Also

* [Heptio Ark](https://github.com/heptio/ark) does sophisticated clusters backups, including volumes
* [Stash](https://github.com/appscode/stash) backups volumes
* [etcd backup operator](https://coreos.com/operators/etcd/docs/latest/user/walkthrough/backup-operator.html)

