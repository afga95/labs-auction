
FROM golang:1.23 as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download


COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o auction cmd/auction/main.go


FROM scratch

WORKDIR /app

COPY --from=builder /app/auction .


EXPOSE 8080


ENTRYPOINT ["./auction"]