# Use a  golang alpine as the base image
FROM public.ecr.aws/docker/library/golang:1.21.0-alpine3.17 as go_builder
RUN apk update
RUN apk add make cmake git alpine-sdk

# Setup

# Read arguments
ARG ARCH
ARG BUILD_DATE
ARG GIT_SHA
ARG SERVICE

# Set env variables
ENV arch=$ARCH
ENV build_date=$BUILD_DATE
ENV commit_hash=$GIT_SHA
ENV service_name=$SERVICE
RUN echo "building service: ${service_name}, build date: ${build_date}, commit hash: ${commit_hash}, architecture: ${arch}"

# Set the working directory
WORKDIR /
COPY . .

# Download dependencies
RUN go mod download


RUN GO111MODULES=on \
    go build \
    -ldflags "-w -linkmode external -extldflags '-static' -X 'github.com/ATMackay/psql-ledger/service.buildDate=${build_date}' -X 'github.com/ATMackay/psql-ledger/service.gitCommit=${commit_hash}'"  \
    -tags netgo \
    -o /out/psqlledger ./cmd/psqlledger/

# Create linux svcuser
RUN mkdir /out/etc && \
    echo "svcuser:x:1010:1010::/sbin/nologin:/bin/false" > /out/etc/passwd && \
    echo "macuser:x:501:20::/sbin/nologin:/bin/false" >> /out/etc/passwd && \
    echo "linuxuser:x:1000:1000::/sbin/nologin:/bin/false" >> /out/etc/passwd && \
    echo "root:x:0:0:root:/sbin/nologin:/bin/false" >> /out/etc/passwd && \
    echo "svcgroup:x:1010:svcuser" > /out/etc/group && \
    echo "macgroup:x:20:macuser" >> /out/etc/group && \
    echo "linuxgroup:x:1000:linuxuser" >> /out/etc/group && \
    echo "root:x:0:root" >> /out/etc/group && \
    mkdir /out/config && \
    chown -R 1010:1010 /out/config


############################################################################################################

#SSL certs
FROM alpine as certs
RUN apk add --no-cache ca-certificates

############################################################################################################


# Copy binary to a scratch container. Let's keep our images nice and small!
FROM scratch
COPY --from=go_builder /out .
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
# Set User
USER svcuser
# Expose the port your application will run on
EXPOSE 8080

# Run the binary
ENTRYPOINT ["/psqlledger"]

