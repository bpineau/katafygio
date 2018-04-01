# katafygio

[![Build Status](https://travis-ci.org/bpineau/katafygio.svg?branch=master)](https://travis-ci.org/bpineau/katafygio)
[![Go Report Card](https://goreportcard.com/badge/github.com/bpineau/katafygio)](https://goreportcard.com/report/github.com/bpineau/katafygio)

katafygio discovers Kubernetes objects (deployments, services, ...),
and continuously save them as yaml files in a git repository.
This provides both real time, continuous backups, and changes history.

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
# (already archived), secrets to keep them confidential, and the leader-elector
# configmap that has low value and changes a lot, causing commits churn.

katafygio \
  -g https://user:token@github.com/myorg/myrepos.git -e /tmp/kfdump \
  -x secret -x pod -x replicaset -y configmap:kube-system/leader-elector
```

## Supported resources

katafygio will discover and save most Kubernetes objects (except CRD).
The supported objects kinds are:

* clusterrolebinding
* configmap
* controllers.go
* cronjob
* daemonset
* deployment
* horizontalpodautoscaler
* ingress
* job
* limitrange
* namespace
* networkpolicy
* persistentvolume
* persistentvolumeclaim
* pod
* poddisruptionbudget
* podsecuritypolicy
* podtemplate
* replicaset
* replicationcontroller
* resourcequota
* rolebinding
* secret
* service
* serviceaccount
* statefulset
* storageclass


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
  -P, --healthcheck-port int         Port for answering healthchecks
  -h, --help                         help for katafygio
  -k, --kube-config string           Kubernetes config path
  -e, --local-dir string             Where to dump yaml files (default "./kubernetes-backup")
  -v, --log-level string             Log level (default "info")
  -o, --log-output string            Log output (default "stderr")
  -r, --log-server string            Log server (if using syslog)
  -i, --resync-interval int          Resync interval in seconds (0 to disable) (default 300)
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

