# Deploy K8S Tool

[![Lint and Testing](https://github.com/appleboy/deploy-k8s/actions/workflows/testing.yml/badge.svg?branch=main)](https://github.com/appleboy/deploy-k8s/actions/workflows/testing.yml)

Generate a Kubeconfig or creating & updating K8s Deployments.

## Installation

Download the latest binary from [release page][1] or install from homebrew.

```sh
brew install appleboy/tap/deploy-k8s
```

[1]: https://github.com/appleboy/deploy-k8s/releases

## Usage

```sh
deploy-k8s --help
```

| Parameter           | Description                                                   | Environment Variables                       |
|---------------------|---------------------------------------------------------------|---------------------------------------------|
| --server            | Address of the Kubernetes cluster `https://hostname:port`      | $PLUGIN_SERVER, $INPUT_SERVER               |
| --skip-tls          | Skip validity check for server's certificate (default: false)   | $PLUGIN_SKIP_TLS_VERIFY, $INPUT_SKIP_TLS_VERIFY |
| --ca-cert           | PEM-encoded certificate authority certificates                 | $PLUGIN_CA_CERT, $INPUT_CA_CERT             |
| --token             | Kubernetes service account token                               | $PLUGIN_TOKEN, $INPUT_TOKEN                 |
| --namespace         | Kubernetes namespace                                           | $PLUGIN_NAMESPACE, $INPUT_NAMESPACE         |
| --proxy-url         | URLs with http, https, and socks5                              | $PLUGIN_PROXY_URL, $INPUT_PROXY_URL         |
| --templates         | Template files, supports glob pattern                          | $PLUGIN_TEMPLATES, $INPUT_TEMPLATES         |
| --output            | Generate Kubernetes config file                                | $PLUGIN_OUTPUT, $INPUT_OUTPUT               |
| --cluster-name      | Cluster name (default: "default")                              | $PLUGIN_CLUSTER_NAME, $INPUT_CLUSTER_NAME   |
| --authinfo-name     | AuthInfo name (default: "default")                             | $PLUGIN_AUTHINFO_NAME, $INPUT_AUTHINFO_NAME |
| --context-name      | Context name (default: "default")                              | $PLUGIN_CONTEXT_NAME, $INPUT_CONTEXT_NAME   |
| --debug             | Enable debug mode (default: false)                             | $PLUGIN_DEBUG, $INPUT_DEBUG                 |
| --help, -h          | Show help                                                     |                                             |
| --version, -v       | Print the version                                             |                                             |

## How To Get Kubernetes Cluster URL

```sh
kubectl config view --raw --minify --flatten \
  -o jsonpath='{.clusters[].cluster.server}'
```

## How To Get Kubernetes CA Certificate

Using Your Own Kubeconfig and don't base64 decode the certificate data.

```sh
kubectl config view --raw --minify --flatten \
  -o jsonpath='{.clusters[].cluster.certificate-authority-data}'
```

## How To Get Kubernetes Token

Create a service account and secret.

```yaml
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: drone-ci
  namespace: default
---
apiVersion: v1
kind: Secret
metadata:
  name: drone-ci
  namespace: default
  annotations:
    kubernetes.io/service-account.name: drone-ci
type: kubernetes.io/service-account-token
```

Get the token from the secret with `default` namespace.

```sh
kubectl get secret drone-ci -n default \
  -o jsonpath='{.data.token}' | base64 -d
```

## How To Get Kubernetes Namespace

```sh
kubectl config view --raw --minify --flatten \
  -o jsonpath='{.contexts[].context.namespace}'
```
