FROM alpine:3.24.1

ARG TARGETOS
ARG TARGETARCH

LABEL maintainer="Bo-Yi Wu <appleboy.tw@gmail.com>" \
  org.label-schema.name="Deploy K8S Tool" \
  org.label-schema.vendor="Bo-Yi Wu" \
  org.label-schema.schema-version="1.0"

LABEL org.opencontainers.image.source=https://github.com/appleboy/deploy-k8s
LABEL org.opencontainers.image.description="Generate a Kubeconfig or creating & updating K8s Deployments."
LABEL org.opencontainers.image.licenses=MIT

RUN apk upgrade --no-cache && \
  apk add --no-cache ca-certificates && \
  addgroup -S -g 10001 deploy && \
  adduser -S -D -u 10001 -G deploy -h /home/deploy deploy && \
  mkdir -p /app && chown deploy:deploy /app

COPY release/${TARGETOS}/${TARGETARCH}/deploy-k8s /bin/

USER 10001:10001

WORKDIR /app

ENTRYPOINT ["/bin/deploy-k8s"]
