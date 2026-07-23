FROM golang:1.26 AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 go build -o /out/server ./cmd
RUN CGO_ENABLED=0 go build -o /out/testclient ./cmd/testclient

FROM gcr.io/distroless/static-debian12 AS server
COPY --from=build /out/server /server
ENTRYPOINT ["/server"]

FROM gcr.io/distroless/static-debian12 AS client
COPY --from=build /out/testclient /testclient
ENTRYPOINT ["/testclient"]
