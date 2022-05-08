FROM golang:1.18 as base

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
RUN CGO_ENABLED=1 GOOS=linux go build -o /main -a -ldflags '-linkmode external -extldflags "-static"' .

FROM scratch
COPY --from=base /main /main
EXPOSE 8080

CMD ["/main"]
