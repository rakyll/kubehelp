# kubehelp

A command line tool that generates kubectl commands from a text
description by using Claude 3 Opus.

## Setup

```sh
$ export ANTHROPIC_API_KEY=<insert key>
$ go install github.com/rakyll/kubehelp/cmd/kubehelp@latest
```

## Usage

```sh
$ kubehelp deploy helloworld to the test namespace and expose it at 8080
kubectl create namespace test
kubectl create deployment helloworld --image=helloworld --namespace test
kubectl expose deployment helloworld --port=8080 --namespace test
Execute: y
namespace/test created
deployment.apps/helloworld created
service/helloworld exposed

$ kubehelp tail the logs from helloworld
kubectl logs -f deployment/helloworld -n test
Execute: y
2024/03/23 00:21:42 Server listening on port 8080

$ kubehelp delete the test namespace
kubectl delete namespace test
Execute: y
namespace "test" deleted
```
