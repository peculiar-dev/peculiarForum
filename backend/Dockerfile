FROM ubuntu:24.04


RUN apt-get update && apt-get install -y curl
RUN apt install -y gcc

WORKDIR /usr/src/app

# Copy the rest of the files into the image.

COPY . .

#install golang
COPY go1.22.0.linux-amd64.tar.gz ./ 

RUN  tar -C /usr/local/ -xzf go1.22.0.linux-amd64.tar.gz

# Define environment variables
ENV PATH="${PATH}:/usr/local/go/bin"
ENV PATH="/go/bin:${PATH}"

#define healthchecks
HEALTHCHECK --start-period=2m --interval=5m --timeout=1m \
   CMD curl -f --retry 6 --max-time 5 --retry-delay 2 --retry-max-time 30 "http://localhost:8080/healthcheck" || bash -c 'kill -s 15 -1 && (sleep 10; kill -s 9 -1)'

RUN go get github.com/mattn/go-sqlite3
RUN go install github.com/mattn/go-sqlite3

# CMD ["CGO_ENABLED=1","go", "run", "main.go"]

CMD CGO_ENABLED=1 go run main.go

