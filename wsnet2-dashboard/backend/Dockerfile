FROM node:lts-slim

RUN apt-get update && apt-get install -y wget make && \ 
  apt-get clean && rm -rf /var/lib/apt/lists/* && \
  wget -O arelo.tar.gz https://github.com/makiuchi-d/arelo/releases/download/v1.11.0/arelo_linux_$([ $(uname -m) = aarch64 ] && echo -n arm64 || echo -n x86_64).tar.gz && \
  tar xf arelo.tar.gz arelo && mv arelo /usr/local/bin && rm arelo.tar.gz && \
  wget -O /wait-for-it.sh https://github.com/vishnubob/wait-for-it/raw/master/wait-for-it.sh > /dev/null 2>&1 && \
  chmod +x /wait-for-it.sh

WORKDIR /repo/wsnet2-dashboard/backend
