FROM golang:latest as builder
RUN go env -w GO111MODULE=on
RUN mkdir /app
WORKDIR /app
COPY . .
RUN go mod tidy -v
RUN go build -o metrics main.go

FROM ubuntu:latest as ubuntu_metrics
RUN mkdir /app
COPY --from=builder /app/metrics /app
CMD ["/app/metrics"]
