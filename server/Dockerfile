FROM golang:1.16

RUN wget -O /wait-for-it.sh https://github.com/vishnubob/wait-for-it/raw/master/wait-for-it.sh > /dev/null 2>&1 && \
    chmod +x /wait-for-it.sh

RUN apt-get update && apt-get install -y unzip && \
    apt-get clean && rm -rf /var/lib/apt/lists/* && \
    wget -O protoc.zip https://github.com/protocolbuffers/protobuf/releases/download/v3.11.4/protoc-3.11.4-linux-x86_64.zip && \
    unzip -o protoc.zip "bin/*" "include/*" -d /usr/local && \
    rm protoc.zip

RUN go install github.com/makiuchi-d/arelo@latest
