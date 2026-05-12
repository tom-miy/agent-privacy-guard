FROM golang:1.24-alpine AS build
WORKDIR /src
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN go build -o /agent-privacy-guard ./cmd/agent-privacy-guard

FROM alpine:3.20
COPY --from=build /agent-privacy-guard /usr/local/bin/agent-privacy-guard
ENTRYPOINT ["agent-privacy-guard"]
