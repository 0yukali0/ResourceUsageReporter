# syntax=docker/dockerfile:1
FROM golang:1.26.0-alpine3.23
RUN apk update && apk --no-cache add build-base
WORKDIR /src
COPY . .
RUN CGO_ENABLED=1  go build -ldflags "-s -w" -o /bin/monitor ./main.go

FROM node:25.6.1-alpine3.23
COPY console /app
WORKDIR /app/nodemonitor
RUN npm run build

FROM alpine:3.23.3
#FROM ubuntu:24.04
#FROM scratch
COPY --from=0 /bin/monitor /bin/monitor
COPY --from=1 /app/nodemonitor/out  /dist
ENV APP_PORT=8080
ENV WEB_DIST=/dist
CMD ["/bin/monitor"]