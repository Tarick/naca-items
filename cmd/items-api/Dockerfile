FROM golang:1.15.5-alpine3.12 as build
WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN GOOS=linux GOARCH=amd64 go mod download
ENV PACKAGE=github.com/Tarick/naca-items
ARG BUILD_VERSION
ADD . .
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -tags 'netgo static_build' -ldflags "-extldflags=-static -w -s -X ${PACKAGE}/version.Version=$BUILD_VERSION -X ${PACKAGE}/version.BuildTime=$(date --utc  +%FT%TZ)" -o ./build/items-api ./cmd/items-api

FROM scratch as final
WORKDIR /
ENV TZ=UTC
COPY --from=build /app/build/items-api /
ENTRYPOINT ["./items-api"]
CMD [ "--config", "./config.yaml" ]
