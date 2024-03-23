# kubehelp [![Go](https://github.com/rakyll/kubehelp/actions/workflows/go.yml/badge.svg)](https://github.com/rakyll/kubehelp/actions/workflows/go.yml)

A command line tool that generates kubectl commands from a text
description by using Claude 3 Opus.

## Setup

```sh
$ export ANTHROPIC_API_KEY=<insert key>
$ go install github.com/rakyll/kubehelp/cmd/kubehelp@latest
```

## Examples

Create a new deployment and expose it:

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

Look up for a service by cluster IP and delete:

```sh
$ kubectl get services -A
NAMESPACE        NAME                   TYPE           CLUSTER-IP       EXTERNAL-IP      PORT(S)            AGE
default          helloworld             LoadBalancer   34.118.225.124   35.222.189.233   9090:30484/TCP     72m
default          kubernetes             ClusterIP      34.118.224.1     <none>           443/TCP            19h
gke-gmp-system   alertmanager           ClusterIP      None             <none>           9093/TCP           19h
gke-gmp-system   gmp-operator           ClusterIP      34.118.235.62    <none>           8443/TCP,443/TCP   19h
kube-system      antrea                 ClusterIP      34.118.236.209   <none>           443/TCP            19h
kube-system      default-http-backend   NodePort       34.118.226.244   <none>           80:30411/TCP       19h
kube-system      kube-dns               ClusterIP      34.118.224.10    <none>           53/UDP,53/TCP      19h
kube-system      metrics-server         ClusterIP      34.118.225.3     <none>           443/TCP            19h

$ kubehelp delete the service listening at 35.222.189.233
kubectl delete service helloworld
Execute: y
service "helloworld" deleted