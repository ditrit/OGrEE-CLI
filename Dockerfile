#
# Dockerfile for the CLI
#
#LABEL author="Ziad Khalaf"
FROM golang:1.19.6-bullseye AS builder
USER root

#Setup app files
WORKDIR /home
ADD . /home/

#Setup build dependencies
RUN go install modernc.org/goyacc@latest
RUN go install github.com/blynn/nex@latest
RUN go get -u github.com/chzyer/test
RUN go get -u golang.org/x/sys

#Generate Binary
RUN make

#Final output image
FROM gcr.io/distroless/base-debian11
WORKDIR /home
ADD . /home/
COPY --from=builder /home/cli /home/
ENTRYPOINT ["/home/cli"]