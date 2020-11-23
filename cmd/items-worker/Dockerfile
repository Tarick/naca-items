FROM golang:1.15.5-alpine3.12 as build
RUN apk --no-cache add git
WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN GOOS=linux GOARCH=amd64 go mod download
ENV PACKAGE=items
ARG BUILD_VERSION
ADD . .
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -tags 'netgo static_build' -ldflags "-extldflags=-static -w -s -X ${PACKAGE}/internal/version.Version=$BUILD_VERSION -X ${PACKAGE}/internal/version.BuildTime=$(date --utc  +%FT%TZ)" -o ./build/items-worker ./cmd/items-worker

FROM scratch as final
WORKDIR /
COPY --from=build /app/build/items-worker /
ENTRYPOINT ["./items-worker"]
