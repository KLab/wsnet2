FROM node:22.11.0-bookworm-slim

RUN apt-get update && apt-get install -y wget make unzip && \
  apt-get clean && rm -rf /var/lib/apt/lists/* && \
  wget -O /wait-for-it.sh https://github.com/vishnubob/wait-for-it/raw/master/wait-for-it.sh > /dev/null 2>&1 && \
  chmod +x /wait-for-it.sh

ARG go_ver="1.23.4"
ARG protoc_ver="29.2"

ENV PATH=/root/go/bin:/go/bin:$PATH

RUN cd / && \
  wget -q -O go.tar.gz https://go.dev/dl/go${go_ver}.linux-$([ $(uname -m) = aarch64 ] && echo -n arm64 || echo -n amd64).tar.gz && \
  tar xf go.tar.gz && \
  rm go.tar.gz && \
  go install github.com/makiuchi-d/arelo@latest && \
  wget -O protoc.zip https://github.com/protocolbuffers/protobuf/releases/download/v${protoc_ver}/protoc-${protoc_ver}-linux-$(if [ $(uname -m) = "aarch64" ]; then echo aarch_64; else echo x86_64; fi).zip && \
  unzip -o protoc.zip "bin/*" "include/*" -d /usr/local && \
  rm protoc.zip

WORKDIR /repo/wsnet2-dashboard/backend
