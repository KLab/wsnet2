FROM node:18.18.2-bullseye-slim

RUN apt-get update && apt-get install -y wget make && \ 
  apt-get clean && rm -rf /var/lib/apt/lists/* && \
  wget -O arelo.tar.gz https://github.com/makiuchi-d/arelo/releases/download/v1.11.0/arelo_linux_$([ $(uname -m) = aarch64 ] && echo -n arm64 || echo -n x86_64).tar.gz && \
  tar xf arelo.tar.gz arelo && mv arelo /usr/local/bin && rm arelo.tar.gz && \
  wget -O /wait-for-it.sh https://github.com/vishnubob/wait-for-it/raw/master/wait-for-it.sh > /dev/null 2>&1 && \
  chmod +x /wait-for-it.sh

ENV PATH /root/go/bin:/go/bin:$PATH
RUN cd / && \
  wget -O go.tar.gz https://go.dev/dl/go1.19.13.linux-$([ $(uname -m) = aarch64 ] && echo -n arm64 || echo -n amd64).tar.gz && \
  tar xvf go.tar.gz && \
  rm go.tar.gz && \
  go install github.com/gopherjs/gopherjs@v1.19.0-beta1

WORKDIR /repo/wsnet2-dashboard/backend
