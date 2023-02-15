# Build Stage
FROM golang:1.12-stretch AS build-stage

LABEL app="build-proksi"
LABEL REPO="https://github.com/leosocy/proksi"

ENV GO111MODULE=on
ENV PROJPATH=/go/src/github.com/leosocy/proksi

# Because of https://github.com/docker/docker/issues/14914
ENV PATH=$PATH:$GOROOT/bin:$GOPATH/bin

ADD . /go/src/github.com/leosocy/proksi
WORKDIR /go/src/github.com/leosocy/proksi

RUN make build-alpine

# Final Stage
FROM alpine:latest

ARG GIT_COMMIT
ARG VERSION
LABEL REPO="https://github.com/leosocy/proksi"
LABEL GIT_COMMIT=$GIT_COMMIT
LABEL VERSION=$VERSION

# Because of https://github.com/docker/docker/issues/14914
ENV PATH=$PATH:/opt/proksi/bin

WORKDIR /opt/proksi/bin

COPY --from=build-stage /go/src/github.com/leosocy/proksi/bin/proksi /opt/proksi/bin/
RUN chmod +x /opt/proksi/bin/proksi
# Install dumb-init
RUN wget -O /usr/bin/dumb-init https://github.com/Yelp/dumb-init/releases/download/v1.2.2/dumb-init_1.2.2_amd64
RUN chmod +x /usr/bin/dumb-init

# Create appuser
RUN adduser -D -g '' proksi
USER proksi

ENTRYPOINT ["/usr/bin/dumb-init", "--"]

CMD ["/opt/proksi/bin/proksi"]
