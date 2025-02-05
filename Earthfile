VERSION 0.6
FROM alpine

ARG RELEASE_VERSION=0.0.1

ARG IMAGE_REPOSITORY=quay.io/kairos
ARG BASE_IMAGE=$IMAGE_REPOSITORY/opensuse:leap-15.5-core-amd64-generic-v2.4.3
ARG BASE_IMAGE_NAME=$(echo $BASE_IMAGE | grep -o [^/]*: | rev | cut -c2- | rev)
ARG BASE_IMAGE_TAG=$(echo $BASE_IMAGE | grep -o :.* | cut -c2-)
ARG PROVIDER_IMAGE_NAME=nodeadm
ARG NODEADM_VERSION=1.0.0
ARG NODEADM_VERSION_TAG=$(echo $NODEADM_VERSION | sed s/+/-/)

ARG LUET_VERSION=0.35.1
ARG GOLINT_VERSION=v1.61.0
ARG GOLANG_VERSION=1.23

luet:
    FROM quay.io/luet/base:$LUET_VERSION
    SAVE ARTIFACT /usr/bin/luet /luet

build-cosign:
    FROM gcr.io/projectsigstore/cosign:v1.13.1
    SAVE ARTIFACT /ko-app/cosign cosign

go-deps:
    FROM us-docker.pkg.dev/palette-images/build-base-images/golang:${GOLANG_VERSION}-alpine
    WORKDIR /build
    COPY go.mod go.sum ./
    RUN go mod download
    RUN apk update
    SAVE ARTIFACT go.mod AS LOCAL go.mod
    SAVE ARTIFACT go.sum AS LOCAL go.sum

BUILD_GOLANG:
    COMMAND
    WORKDIR /build
    COPY . ./
    ARG BIN
    ARG SRC
    RUN go-build-static.sh -a -o ${BIN} ./${SRC}
    SAVE ARTIFACT ${BIN} ${BIN} AS LOCAL build/${BIN}

ARCH:
    COMMAND
    ARG TARGETPLATFORM
    FROM alpine
    RUN echo "$TARGETPLATFORM" | awk -F/ '{print $2}' > ARCH
    SAVE ARTIFACT ARCH ARCH

VERSION:
    COMMAND
    FROM alpine
    RUN apk add git
    COPY . ./
    RUN echo $(git describe --exact-match --tags || echo "v0.0.0-$(git log --oneline -n 1 | cut -d" " -f1)") > VERSION
    SAVE ARTIFACT VERSION VERSION

build-provider:
    FROM +go-deps
    DO +BUILD_GOLANG --BIN=agent-provider-nodeadm --SRC=main.go

build-provider-package:
    DO +VERSION
    ARG VERSION=$(cat VERSION)

    FROM scratch

    COPY +build-provider/agent-provider-nodeadm /system/providers/agent-provider-nodeadm
    COPY scripts/ /opt/nodeadmutil/scripts/

    SAVE IMAGE --push $IMAGE_REPOSITORY/provider-nodeadm:latest
    SAVE IMAGE --push $IMAGE_REPOSITORY/provider-nodeadm:${VERSION}

lint:
    FROM golang:$GOLANG_VERSION
    RUN wget -O- -nv https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s ${GOLINT_VERSION}
    WORKDIR /build
    COPY . .
    RUN golangci-lint run --timeout=5m

test:
    FROM +go-deps
    COPY . ./
    RUN go test -v -timeout 5m ./...

DOWNLOAD_BINARIES:
    COMMAND
    ARG --required ARCH
    RUN curl -L --remote-name-all https://hybrid-assets.eks.amazonaws.com/releases/v${NODEADM_VERSION}/bin/linux/${ARCH}/nodeadm

SAVE_IMAGE:
    COMMAND
    ARG VERSION
    SAVE IMAGE --push $IMAGE_REPOSITORY/${BASE_IMAGE_NAME}-${PROVIDER_IMAGE_NAME}:${NODEADM_VERSION_TAG}
    SAVE IMAGE --push $IMAGE_REPOSITORY/${BASE_IMAGE_NAME}-${PROVIDER_IMAGE_NAME}:${NODEADM_VERSION_TAG}_${VERSION}

docker:
    DO +ARCH
    ARG ARCH=$(cat ARCH)

    DO +VERSION
    ARG VERSION=$(cat VERSION)

    FROM $BASE_IMAGE

    WORKDIR /usr/bin

    DO +DOWNLOAD_BINARIES --ARCH=$ARCH

    RUN chmod +x nodeadm

    COPY +luet/luet /usr/bin/luet

    WORKDIR /

    ENV OS_ID=${BASE_IMAGE_NAME}-nodeadm
    ENV OS_NAME=$OS_ID:${BASE_IMAGE_TAG}
    ENV OS_REPO=${IMAGE_REPOSITORY}
    ENV OS_VERSION=${NODEADM_VERSION_TAG}_${VERSION}
    ENV OS_LABEL=${BASE_IMAGE_TAG}_${NODEADM_VERSION_TAG}_${VERSION}
    RUN envsubst >>/etc/os-release </usr/lib/os-release.tmpl

    RUN mkdir -p /opt/nodeadmutil/scripts
    COPY scripts/* /opt/nodeadmutil/scripts/

    COPY +build-provider/agent-provider-nodeadm /system/providers/agent-provider-nodeadm

    DO +SAVE_IMAGE --VERSION=$VERSION

cosign:
    ARG --required ACTIONS_ID_TOKEN_REQUEST_TOKEN
    ARG --required ACTIONS_ID_TOKEN_REQUEST_URL

    ARG --required REGISTRY
    ARG --required REGISTRY_USER
    ARG --required REGISTRY_PASSWORD

    DO +VERSION
    ARG VERSION=$(cat VERSION)

    FROM docker

    ENV ACTIONS_ID_TOKEN_REQUEST_TOKEN=${ACTIONS_ID_TOKEN_REQUEST_TOKEN}
    ENV ACTIONS_ID_TOKEN_REQUEST_URL=${ACTIONS_ID_TOKEN_REQUEST_URL}

    ENV REGISTRY=${REGISTRY}
    ENV REGISTRY_USER=${REGISTRY_USER}
    ENV REGISTRY_PASSWORD=${REGISTRY_PASSWORD}

    ENV COSIGN_EXPERIMENTAL=1
    COPY +build-cosign/cosign /usr/local/bin/

    RUN echo $REGISTRY_PASSWORD | docker login -u $REGISTRY_USER --password-stdin $REGISTRY

    DO +SAVE_IMAGE --VERSION=$VERSION

docker-all-platforms:
    BUILD --platform=linux/amd64 +docker

provider-package-all-platforms:
    BUILD --platform=linux/amd64 +build-provider-package
    BUILD --platform=linux/arm64 +build-provider-package

cosign-all-platforms:
    BUILD --platform=linux/amd64 +cosign
