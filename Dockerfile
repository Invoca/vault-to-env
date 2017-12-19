FROM golang:1.9
WORKDIR /go/src/github.com/invoca/vault-to-env
COPY . .
RUN go get github.com/jarcoal/httpmock \
    && CGO_ENABLED=0 GOOS=linux go build

FROM scratch
COPY --from=0 /go/src/github.com/invoca/vault-to-env/vault-to-env /
ENTRYPOINT ["/vault-to-env"]
CMD ["-h"]
