# syntax=docker/dockerfile:1
FROM golang:1.25
WORKDIR /src
COPY . .
RUN go build -ldflags "-s -w" -o /bin/monitor ./main.go

FROM ubuntu:26.04
COPY --from=0 /bin/monitor /bin/monitor
ENV APP_PORT=8080
CMD ["/bin/monitor"]