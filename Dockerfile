# syntax=docker/dockerfile:1
FROM golang:1.19-alpine AS build
WORKDIR /app/my-golibs
COPY . .
RUN go mod tidy
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -tags=nomsgpack -o /my-golibs .

FROM scratch
WORKDIR /
COPY --from=build /my-golibs /my-golibs
EXPOSE 8045
ENTRYPOINT ["/my-golibs"]