FROM golang:1.18.2 as builder
ENV GOOS linux
ENV CGO_ENABLED 0
WORKDIR /app
COPY . .
RUN go build -o app

FROM alpine:3.16.0 as production
COPY --from=builder app .
EXPOSE 80
CMD ./app