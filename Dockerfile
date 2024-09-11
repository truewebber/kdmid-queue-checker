FROM golang:1.23.1-bookworm as builder

WORKDIR /app

COPY . .

RUN PWGO_VER=$(grep -oE "playwright-go v\S+" /app/go.mod | sed 's/playwright-go //g') \
    && go install github.com/playwright-community/playwright-go/cmd/playwright@${PWGO_VER}

RUN go build -o /app/bin/checker ./cmd/checker

FROM debian:bookworm

WORKDIR /home/checker

COPY --from=builder --chown=checker /go/bin/playwright ./

RUN apt-get update && apt-get install -y ca-certificates tzdata \
    && ./playwright install --with-deps \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder --chown=checker /app/bin/checker ./

CMD '/home/checker/checker'
