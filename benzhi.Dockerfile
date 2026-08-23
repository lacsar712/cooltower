FROM golang:1.22 AS build
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build -o /out/cooltower ./cmd/cooltower

FROM gcr.io/distroless/static-debian12
COPY --from=build /out/cooltower /cooltower
ENTRYPOINT ["/cooltower"]
