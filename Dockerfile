### Stage 1 - Build
FROM golang:1.14-alpine as build
WORKDIR /go/src/github.com/mdblp/gatekeeper
RUN apk --no-cache update && \
    apk --no-cache upgrade && \
    apk --no-cache add git make tzdata ca-certificates

COPY . .
RUN make build

### Stage 2 - Serve production-ready release
FROM alpine:3.12 as production
RUN apk --no-cache update && \
    apk --no-cache upgrade && \
    apk --no-cache add tzdata ca-certificates && \
    adduser -D gatekeeper
WORKDIR /app
COPY --from=build --chown=root:root /go/src/github.com/mdblp/gatekeeper/gatekeeper /app
USER gatekeeper
ENTRYPOINT ["/app/gatekeeper"]
