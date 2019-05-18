# Build Stage
FROM golang:1.12-stretch AS build-stage

LABEL app="build-intelliproxy"
LABEL REPO="https://github.com/Leosocy/IntelliProxy"

ENV GO111MODULE=on
ENV PROJPATH=/go/src/github.com/Leosocy/IntelliProxy

# Because of https://github.com/docker/docker/issues/14914
ENV PATH=$PATH:$GOROOT/bin:$GOPATH/bin

ADD . /go/src/github.com/Leosocy/IntelliProxy
WORKDIR /go/src/github.com/Leosocy/IntelliProxy

RUN make build-alpine

# Final Stage
FROM alpine:latest

ARG GIT_COMMIT
ARG VERSION
LABEL REPO="https://github.com/Leosocy/IntelliProxy"
LABEL GIT_COMMIT=$GIT_COMMIT
LABEL VERSION=$VERSION

# Because of https://github.com/docker/docker/issues/14914
ENV PATH=$PATH:/opt/intelliproxy/bin

WORKDIR /opt/intelliproxy/bin

COPY --from=build-stage /go/src/github.com/Leosocy/IntelliProxy/bin/intelliproxy /opt/intelliproxy/bin/
RUN chmod +x /opt/intelliproxy/bin/intelliproxy
# Install dumb-init
RUN wget -O /usr/bin/dumb-init https://github.com/Yelp/dumb-init/releases/download/v1.2.2/dumb-init_1.2.2_amd64
RUN chmod +x /usr/bin/dumb-init

# Create appuser
RUN adduser -D -g '' intelliproxy
USER intelliproxy

ENTRYPOINT ["/usr/bin/dumb-init", "--"]

CMD ["/opt/intelliproxy/bin/intelliproxy"]
