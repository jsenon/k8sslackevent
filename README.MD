# K8S Event Watcher

[![CircleCI](https://circleci.com/gh/jsenon/demo-istio.svg?style=svg)](https://circleci.com/gh/jsenon/k8sslackevent)
[![Go Report Card](https://goreportcard.com/badge/github.com/jsenon/demo-istio)](https://goreportcard.com/report/github.com/jsenon/k8sslackevent)

An event watcher for Kubernetes Environment, and publish on slack

## Func

- Pod Event Watcher on default and kube-system namespace
- Node Event Watcher
- Event Watcher
- Publisher on slack
- Watch OOMKilled Event and retrieve pod name
- External or Internal k8s api

### Build

### Usage

Right now pod watcher is available on default and kube-system namespaces

Create Webhook slack application and export URL

```sh
export SLACK_URL=https://hooks.slack.com/xxx
```

For internal k8s:

```sh
k8sslackevent serve --api internal
```

For External k8s, its based on current context in your .kube/config file:

```sh
k8sslackevent serve --api external
```

### Env

- `SLACK_URL` = Slack URL Hook

### TIPS

If you receive `cannot execute binary file: Exec format error`

Compile with  `CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build main.go`

And build `docker build -t jsenon/k8sslackevent:vx.x .`

### TODO

- [ ] Specify Pod watcher namespace as flag
- [ ] Add Event bus integration