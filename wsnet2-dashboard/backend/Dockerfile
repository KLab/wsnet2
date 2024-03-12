FROM node:18.18.2-bullseye-slim

RUN apt-get update && apt-get install -y wget make unzip && \
  apt-get clean && rm -rf /var/lib/apt/lists/* && \
  wget -O /wait-for-it.sh https://github.com/vishnubob/wait-for-it/raw/master/wait-for-it.sh > /dev/null 2>&1 && \
  chmod +x /wait-for-it.sh

ARG go_ver="1.22.0"
ARG gopherjs_ver="v1.19.0-beta1"
ARG gopherjs_go="go1.19.13"
ARG protoc_ver="24.2"

ENV PATH=/root/go/bin:/go/bin:$PATH
ENV GOPHERJS_GOROOT=/root/sdk/${gopherjs_go}

RUN cd / && \
  wget -O go.tar.gz https://go.dev/dl/go${go_ver}.linux-$([ $(uname -m) = aarch64 ] && echo -n arm64 || echo -n amd64).tar.gz && \
  tar xvf go.tar.gz && \
  rm go.tar.gz && \
  go install github.com/makiuchi-d/arelo@latest && \
  go install github.com/gopherjs/gopherjs@${gopherjs_ver} && \
  go install golang.org/dl/${gopherjs_go}@latest && \
  $gopherjs_go download && \
  wget -O protoc.zip https://github.com/protocolbuffers/protobuf/releases/download/v${protoc_ver}/protoc-${protoc_ver}-linux-$(if [ $(uname -m) = "aarch64" ]; then echo aarch_64; else echo x86_64; fi).zip && \
  unzip -o protoc.zip "bin/*" "include/*" -d /usr/local && \
  rm protoc.zip

WORKDIR /repo/wsnet2-dashboard/backend
