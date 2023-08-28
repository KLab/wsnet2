FROM golang:1.21

ARG protoc_ver="24.2"

RUN apt-get update && apt-get install -y unzip && \
    apt-get clean && rm -rf /var/lib/apt/lists/* && \
    wget -O protoc.zip https://github.com/protocolbuffers/protobuf/releases/download/v${protoc_ver}/protoc-${protoc_ver}-linux-$(if [ $(uname -m) = "aarch64" ]; then echo aarch_64; else echo x86_64; fi).zip && \
    unzip -o protoc.zip "bin/*" "include/*" -d /usr/local && \
    rm protoc.zip

RUN go install github.com/makiuchi-d/arelo@latest && \
    git config --global --add safe.directory /repo
