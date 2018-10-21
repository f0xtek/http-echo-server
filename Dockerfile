FROM golang:1.11-alpine as builder
WORKDIR /build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a --installsuffix cgo -o http-echo *.go

FROM scratch
COPY --from=builder /build/http-echo .
CMD ["./http-echo"]