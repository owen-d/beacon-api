# Start from a Debian image with the latest version of Go installed
# and a workspace (GOPATH) configured at /go.
FROM golang:1.8.3

ENV REPO github.com/owen-d/beacon-api/
MAINTAINER "ow.diehl@gmail.com"

# Copy the local package files to the container's workspace.
ADD . /go/src/${REPO}

RUN go install $REPO
# Run the outyet command by default when the container starts.
ENTRYPOINT /go/bin/beacon-api