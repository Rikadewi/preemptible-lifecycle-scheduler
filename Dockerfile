FROM asia.gcr.io/warung-support/golang-base:latest AS builder

ARG SSH_PRIVATE_KEY

ARG REPO_NAME
ENV APP_DIR=$REPO_NAME
ENV GO111MODULE=on


RUN mkdir -p /app/src/handler
WORKDIR /app/src/handler

# Download public key for gitlab.warungpintar.co
RUN mkdir -p /root/.ssh/ \
    && touch /root/.ssh/config

RUN echo "${SSH_PRIVATE_KEY}" > /root/.ssh/id_rsa \
    && chmod 600 /root/.ssh/id_rsa \
    && echo "IdentityFile /root/.ssh/id_rsa" >> /root/.ssh/config \
    && echo -e "Host *\n\tStrictHostKeyChecking no\n\n" > /root/.ssh/config \
    && git config --global url."git@gitlab.warungpintar.co:".insteadOf "https://gitlab.warungpintar.co/"

# manage dependencies
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /app/src/handler/preemptible-lifecycle-scheduler /app/src/handler/main.go
#RUN go test $(go list ./... | grep -v /vendor/) -cover

#remove ssh key
RUN rm -rf /root/.ssh

FROM alpine:3.10

# Setting timezone
ENV TZ=Asia/Jakarta
RUN apk add -U tzdata
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

ARG env_stage_name
ENV ENV_STAGE=${env_stage_name}

  # Add non root user and certs
RUN apk --no-cache add ca-certificates \
  && addgroup -S app && adduser -S -g app app \
  && mkdir -p /home/app/etc/preemptible-lifecycle-scheduler/ \
  && chown -R app /home/app

WORKDIR /home/app
RUN mkdir -p config

COPY --from=builder /app/src/handler/preemptible-lifecycle-scheduler .

USER app

CMD ["./preemptible-lifecycle-scheduler"]
