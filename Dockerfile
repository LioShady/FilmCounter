FROM golang:alpine AS builder
WORKDIR /app
COPY . .
RUN go mod tidy && go build -o app

FROM alpine
WORKDIR /app
COPY --from=builder /app/app .
COPY --from=builder /app/internal/web/templates /app/internal/web/templates
COPY --from=builder /app/.env .
EXPOSE 8080

CMD ["./app"]



