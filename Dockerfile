FROM golang:alpine AS builder

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /build

COPY go.mod go.sum main.go ./

COPY controller ./controller
COPY internal ./internal
COPY model ./model

RUN go mod download

RUN go mod tidy

RUN go build -o main .

WORKDIR /dist

RUN cp /build/main .

FROM scratch

COPY --from=builder /dist/main .

ENV HOST_CLUSTER_NAME=${HOST_CLUSTER_NAME} \
    KARMADA_API=${KARMADA_API} \
    KARMADA_TOKEN=${KARMADA_TOKEN} \
    NATS_ID=${NATS_ID} \
    NATS_PASSWORD=${NATS_PASSWORD} \
    NATS_BUCKET_NAME=${NATS_BUCKET_NAME} \
    NATS_SUBJECT_NAME=${NATS_SUBJECT_NAME} \
    NATS_URL=${NATS_URL} \
    VAULT_ROLE_ID=${VAULT_ROLE_ID} \
    VAULT_ROLE_NAME=${VAULT_ROLE_NAME} \
    VAULT_SECRET_ID=${VAULT_SECRET_ID} \
    VAULT_URL=${VAULT_URL}


ENTRYPOINT ["/main"]
