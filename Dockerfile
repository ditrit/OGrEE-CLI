
FROM golang:1.19.6-bullseye AS builder
USER root

#Setup app files
WORKDIR /home
ADD . /home/

RUN make

#Final output image
FROM busybox:latest
WORKDIR /home
ADD . /home/
COPY --from=builder /home/cli /home/