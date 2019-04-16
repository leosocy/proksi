# Build Stage
FROM golang:1.12-stretch AS build-stage

LABEL app="build-gipp"
LABEL REPO="https://github.com/Leosocy/gipp"

ENV GO111MODULE=on
ENV PROJPATH=/go/src/github.com/Leosocy/gipp

# Because of https://github.com/docker/docker/issues/14914
ENV PATH=$PATH:$GOROOT/bin:$GOPATH/bin

ADD . /go/src/github.com/Leosocy/gipp
WORKDIR /go/src/github.com/Leosocy/gipp

RUN make build-alpine

# Final Stage
FROM alpine:latest

ARG GIT_COMMIT
ARG VERSION
LABEL REPO="https://github.com/Leosocy/gipp"
LABEL GIT_COMMIT=$GIT_COMMIT
LABEL VERSION=$VERSION

# Because of https://github.com/docker/docker/issues/14914
ENV PATH=$PATH:/opt/gipp/bin

WORKDIR /opt/gipp/bin

COPY --from=build-stage /go/src/github.com/Leosocy/gipp/bin/gipp /opt/gipp/bin/
RUN chmod +x /opt/gipp/bin/gipp
# Install dumb-init
RUN wget -O /usr/bin/dumb-init https://github.com/Yelp/dumb-init/releases/download/v1.2.2/dumb-init_1.2.2_amd64
RUN chmod +x /usr/bin/dumb-init

# Create appuser
RUN adduser -D -g '' gipp
USER gipp

ENTRYPOINT ["/usr/bin/dumb-init", "--"]

CMD ["/opt/gipp/bin/gipp"]
