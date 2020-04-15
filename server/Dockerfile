FROM golang:1.14

RUN go get github.com/oxequa/realize

RUN wget -O /wait-for-it.sh https://raw.githubusercontent.com/vishnubob/wait-for-it/master/wait-for-it.sh > /dev/null 2>&1 && \
    chmod +x /wait-for-it.sh
