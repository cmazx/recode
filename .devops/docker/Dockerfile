FROM golang:latest as builder

RUN mkdir /var/app
WORKDIR /var/app
#
RUN apt-get update
RUN apt-get install -q -y git wget build-essential pkg-config \
    libjpeg-dev libpng-dev libtiff-dev \
    libgif-dev libx11-dev libmagickwand-dev --no-install-recommends

COPY go.mod .
COPY go.sum .

RUN go mod download

ADD . .
RUN go build -o /var/app/app ./cmd

FROM cmazx/recode:0.1

#RUN apt-get update
#RUN apt-get install -q -y git ca-certificates wget build-essential pkg-config \
#    libjpeg-dev libpng-dev libtiff-dev \
#    libgif-dev libx11-dev libmagickwand-dev --no-install-recommends
#RUN update-ca-certificates


#RUN mkdir -p /var/app
WORKDIR /var/app
COPY --chown=0:0 --from=builder /var/app/app /var/app/main

# Run the executable
CMD ["./main"]