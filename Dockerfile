# Copyright 2025 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Stage 1: build the Vue web UI. A production build is minified and small
# (~140 KB), and gets embedded into the Go binary in the next stage. This stage
# is discarded (only dist/ is copied forward), so the base image size is
# irrelevant; slim/glibc avoids the musl native-binary pitfalls of alpine.
FROM node:22-slim AS ui
WORKDIR /ui
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

# Stage 2: build the statically-linked Go binary with the UI embedded.
FROM golang:1.26.5 AS builder

ARG TAG
ARG COMMIT

ADD ./internal /go/src/apphub-app-creator/internal
ADD ./cmd /go/src/apphub-app-creator/cmd

COPY .github/workflows/licenses.tpl /go/src/apphub-app-creator
COPY go.mod go.sum /go/src/apphub-app-creator/

# Place the built UI where the cmd package embeds it (internal/cmd/webui_embed.go
# reads internal/cmd/webdist behind the "embedui" build tag).
COPY --from=ui /ui/dist /go/src/apphub-app-creator/internal/cmd/webdist

WORKDIR /go/src/apphub-app-creator

ENV GO111MODULE=on
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags embedui -trimpath -buildvcs=true -a -ldflags='-s -w -extldflags "-static" -X main.version='${TAG}' -X main.commit='${COMMIT}' -X main.date='$(date -u +%FT%TZ) -o /go/bin/apphub-app-creator /go/src/apphub-app-creator/cmd/apphub-app-creator/apphub-app-creator.go
RUN GOBIN=/tmp/ go install github.com/google/go-licenses@v1.6.0
RUN /tmp/go-licenses report ./... --template /go/src/apphub-app-creator/licenses.tpl --ignore internal > /tmp/third-party-licenses.txt 2> /dev/null || echo "Ignore warnings"

# use debug because it includes busybox
FROM gcr.io/distroless/static-debian12:debug-nonroot@sha256:d5563cc7f2f44313f332e91138cc8c6a158899afeeeab2fce3b0f9ccdb3cf9ee
LABEL org.opencontainers.image.url='https://github.com/srinandan/apphub-app-creator' \
    org.opencontainers.image.documentation='https://github.com/srinandan/apphub-app-creator' \
    org.opencontainers.image.source='https://github.com/srinandan/apphub-app-creator' \
    org.opencontainers.image.vendor='Google LLC' \
    org.opencontainers.image.licenses='Apache-2.0' \
    org.opencontainers.image.description='This is a tool to generate App Hub Applications based on CAIS Asset Search'

COPY --from=builder /go/bin/apphub-app-creator /usr/local/bin/apphub-app-creator
COPY --chown=nonroot:nonroot LICENSE.txt /
COPY --from=builder --chown=nonroot:nonroot /tmp/third-party-licenses.txt /

ENTRYPOINT [ "apphub-app-creator" ]
