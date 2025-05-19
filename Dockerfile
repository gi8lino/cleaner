FROM golang:1.24-alpine as builder

WORKDIR /workspace
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GO111MODULE=on go build -ldflags="-w -s" -a -installsuffix nocgo -o /cleaner cmd/cleaner/main.go

FROM gcr.io/distroless/static:nonroot

WORKDIR /
COPY --from=builder /cleaner ./
USER nonroot:nonroot
ENTRYPOINT ["./cleaner"]


