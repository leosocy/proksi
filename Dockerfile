# Build Stage
FROM leosocy/gipp:onbuild:1.11 AS build-stage

LABEL app="build-gipp"
LABEL REPO="https://github.com/Leosocy/gipp"

ENV PROJPATH=/go/src/github.com/Leosocy/gipp

# Because of https://github.com/docker/docker/issues/14914
ENV PATH=$PATH:$GOROOT/bin:$GOPATH/bin

ADD . /go/src/github.com/Leosocy/gipp
WORKDIR /go/src/github.com/Leosocy/gipp

RUN make build-alpine

# Final Stage
FROM leosocy/go-alpine:latest

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

# Create appuser
RUN adduser -D -g '' gipp
USER gipp

ENTRYPOINT ["/usr/bin/dumb-init", "--"]

CMD ["/opt/gipp/bin/gipp"]
