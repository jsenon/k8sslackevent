# Demo Istio

[![CircleCI](https://circleci.com/gh/jsenon/demo-istio.svg?style=svg)](https://circleci.com/gh/jsenon/k8sslackevent)
[![Go Report Card](https://goreportcard.com/badge/github.com/jsenon/demo-istio)](https://goreportcard.com/report/github.com/jsenon/k8sslackevent)

An event watcher for Kubernetes Environment, and publish on slack

## Func

- Pod Event Watcher
- Node Event Watcher
- Event Watcher
- Publisher on slack

### Build

### Usage

Generate Health, answer could be 200 or 5xx if `MY_ANSWER_CODE` is set:

```sh
export SLACK_URL=https://hooks.slack.com/xxx
```

### Env

- `SLACK_URL` = Slack URL Hook

### TIPS

If you receive `cannot execute binary file: Exec format error`

Compile with  `CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build main.go`

And build `docker build -t jsenon/k8sslackevent:vx.x .`