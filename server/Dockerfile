FROM golang:1.14

RUN go get github.com/oxequa/realize

RUN wget -O /wait-for-it.sh https://github.com/vishnubob/wait-for-it/raw/master/wait-for-it.sh > /dev/null 2>&1 && \
    chmod +x /wait-for-it.sh

RUN apt-get update && apt-get install -y unzip

RUN curl -L https://github.com/protocolbuffers/protobuf/releases/download/v3.11.4/protoc-3.11.4-linux-x86_64.zip > protoc.zip && \
    unzip -o protoc.zip -d /usr/local bin/protoc && \
    unzip -o protoc.zip -d /usr/local include/* && \
    rm protoc.zip

RUN go get github.com/golang/protobuf/protoc-gen-go
RUN go get github.com/favadi/protoc-go-inject-tag
