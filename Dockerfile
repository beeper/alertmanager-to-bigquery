FROM golang:1.20-alpine3.17 as development
WORKDIR /build
COPY . /build
RUN go build \
    -o alertmanager-to-bigquery \
    -ldflags "-X main.Commit=$COMMIT_HASH -X 'main.BuildTime=`date '+%b %_d %Y, %H:%M:%S'`'" \
    ./cmd/alertmanager-to-bigquery
RUN go install github.com/mitranim/gow@latest
ENTRYPOINT ["gow", "run", "./cmd/alertmanager-to-bigquery"]


FROM alpine:3.17
RUN apk add --no-cache ca-certificates
COPY --from=development /build/alertmanager-to-bigquery /alertmanager-to-bigquery
ENTRYPOINT ["/alertmanager-to-bigquery"]
