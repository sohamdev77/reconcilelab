FROM golang:1.26 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -o /reconcilelab ./cmd/reconcilelab
FROM scratch
COPY --from=build /reconcilelab /reconcilelab
USER 65532:65532
ENTRYPOINT ["/reconcilelab"]
CMD ["version"]
