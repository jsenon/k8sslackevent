# K8S Event Watcher

[![CircleCI](https://circleci.com/gh/jsenon/demo-istio.svg?style=svg)](https://circleci.com/gh/jsenon/k8sslackevent)
[![Go Report Card](https://goreportcard.com/badge/github.com/jsenon/demo-istio)](https://goreportcard.com/report/github.com/jsenon/k8sslackevent)

An event watcher for Kubernetes Environment, and publish on slack

## Func

- Pod Event Watcher on default namespace
- Node Event Watcher
- Event Watcher
- Publisher on slack
- Watch OOMKilled Event and retrieve pod name

### Build

### Usage

Create Webhook slack application and export URL

```sh
export SLACK_URL=https://hooks.slack.com/xxx
```

Right now pod watcher is only available on default namespace

### Env

- `SLACK_URL` = Slack URL Hook

### TIPS

If you receive `cannot execute binary file: Exec format error`

Compile with  `CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build main.go`

And build `docker build -t jsenon/k8sslackevent:vx.x .`

### TODO

- [ ] Specify Pod watcher namespace as flag
- [ ] Add Event bus integration