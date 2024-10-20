# syntax=docker/dockerfile:1

#FROM mcr.microsoft.com/playwright:v1.43.1-jammy
FROM ubuntu:24.04


#RUN npm config set proxy http://proxy.psu.edu:3128
#RUN npm config set https-proxy http://proxy.psu.edu:3128

#RUN npm --proxy http://proxy.psu.edu:3128  install -g npm@10.1.0

RUN apt-get update && apt-get install -y curl
RUN apt install -y gcc

WORKDIR /usr/src/app

# Copy the rest of the files into the image.
#COPY ./PE_Playwright/ ./
COPY . .

#RUN npm --proxy http://proxy.psu.edu:3128  install
#RUN npx --proxy http://proxy.psu.edu:3128 playwright install

# install dotenv
#RUN npm --proxy http://proxy.psu.edu:3128  install dotenv

#install golang
COPY go1.22.0.linux-amd64.tar.gz ./ 

RUN  tar -C /usr/local/ -xzf go1.22.0.linux-amd64.tar.gz

#COPY package.json package-lock.json ./

#RUN --mount=type=cache,target=/root/.npm \
#    npm i

# Copy the rest of the files into the image.
#COPY /home/bishop/autoTest/PE_Playwright/ ./

# Define environment variables

ENV PATH="${PATH}:/usr/local/go/bin"
ENV PATH="/go/bin:${PATH}"

#define healthchecks
HEALTHCHECK --interval=1m --timeout=1m --start-period=30s \
   CMD curl -f --retry 6 --max-time 5 --retry-delay 1 --retry-max-time 30 "http://localhost:8080/healthcheck" || bash -c 'kill -s 15 -1 && (sleep 10; kill -s 9 -1)'


RUN go get github.com/mattn/go-sqlite3
RUN go install github.com/mattn/go-sqlite3

# CMD ["go", "run", "main.go"]

CMD CGO_ENABLED=1 go run main.go

# Set entry point to shell
#CMD ["/bin/bash"]

