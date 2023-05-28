# Deploy K8S Tool

[![Lint and Testing](https://github.com/appleboy/deploy-k8s/actions/workflows/lint.yml/badge.svg?branch=main)](https://github.com/appleboy/deploy-k8s/actions/workflows/lint.yml)

Generate a Kubeconfig or creating & updating K8s Deployments.

## Installation

Download the latest binary from [release page][1] or install from homebrew.

```sh
brew install appleboy/tap/deploy-k8s
```

[1]: https://github.com/appleboy/deploy-k8s/releases

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
