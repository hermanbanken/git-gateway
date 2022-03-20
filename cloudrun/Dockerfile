# syntax = docker/dockerfile:1-experimental
FROM golang:1.18 as build

WORKDIR /usr/src/app

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build go build -v -o /usr/local/bin/app ./cloudrun

FROM gcr.io/distroless/base-debian11
COPY --from=build /usr/local/bin/app /
CMD ["/app"]
