# Args needs to be renewed for each build stage that uses these args
ARG workDir=/app
ARG workDirRepo=/app/services
ARG GoVersion=1.18

FROM golang:$GoVersion AS build-base

# Renew Arg and set working directory to /app
ARG workDir
WORKDIR $workDir

# Download go installation files
RUN wget --progress=dot:giga https://raw.githubusercontent.com/veraison/services/main/go.mod &&\
    wget --progress=dot:giga https://raw.githubusercontent.com/veraison/services/main/go.sum

# Install build dependencies and tools
RUN apt-get update \
    && DEBIAN_FRONTEND=noninteractive apt-get install \
        --assume-yes \
        --no-install-recommends \
        apt-transport-https \
        apt-utils \
        jq \
        tree \
        protobuf-compiler \
        python3 \
        python3-pip \
        libprotobuf-dev \
        libsqlite3-dev \
        sqlite3 \
        zlib1g-dev \
    && apt-get clean \
    && apt-get autoremove --assume-yes \
    && rm -rf /var/lib/apt/lists/* /var/tmp/* /tmp/*

# Install config templating dependencies
RUN pip install --no-cache-dir Jinja2

# Download Go modules
RUN go mod download &&\
    go install github.com/golang/mock/mockgen@v1.6.0 &&\
    go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.26 &&\
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.1 &&\
    go install github.com/mitchellh/protoc-gen-go-json@v1.1.0 &&\
    go install github.com/veraison/corim/cocli@demo-psa-1.0.0 &&\
    go install github.com/veraison/evcli@demo-psa-1.0.0


FROM build-base as common-build
# Renew build arguments
ARG workDirRepo
ARG PROVISIONING_DEPLOY_PREFIX
ARG VERIFICATION_DEPLOY_PREFIX
ARG VTS_DEPLOY_PREFIX
ARG VTS_PROVISIONING_LOCAL_IP_ADDRESS
ARG VTS_VERIFICATION_LOCAL_IP_ADDRESS
ARG VTS_PROVISIONING_LOCAL_IP_ADDRESS_PORT
ARG VTS_VERIFICATION_LOCAL_IP_ADDRESS_PORT
ARG BIN_DIR
ARG LOG_DIR
ARG PLUGIN_DIR
ARG INPUT_FILE_DIR

# Set envrionment variables for deploy directories and communication ip addresses
ENV PROVISIONING_DEPLOY_PREFIX $workDirRepo$PROVISIONING_DEPLOY_PREFIX
ENV VERIFICATION_DEPLOY_PREFIX $workDirRepo$VERIFICATION_DEPLOY_PREFIX
ENV VTS_DEPLOY_PREFIX $workDirRepo$VTS_DEPLOY_PREFIX
ENV VTS_PROVISIONING_LOCAL_IP_ADDRESS $VTS_PROVISIONING_LOCAL_IP_ADDRESS
ENV VTS_VERIFICATION_LOCAL_IP_ADDRESS $VTS_VERIFICATION_LOCAL_IP_ADDRESS 
ENV VTS_PROVISIONING_LOCAL_IP_ADDRESS_PORT $VTS_PROVISIONING_LOCAL_IP_ADDRESS_PORT
ENV VTS_VERIFICATION_LOCAL_IP_ADDRESS_PORT $VTS_VERIFICATION_LOCAL_IP_ADDRESS_PORT

# Set environment variables for binaries, logs, plugins and input directories
ENV BIN_DIR $BIN_DIR
ENV LOG_DIR $LOG_DIR
ENV PLUGIN_DIR $PLUGIN_DIR
ENV INPUT_FILE_DIR $INPUT_FILE_DIR

# Clone deployment branch from remote (TODO: change this to clone from main once Docker change has been integrated)
RUN git clone --branch deployment https://github.com/veraison/services.git

# Set working directory to repo
WORKDIR $workDirRepo

# Generates a config for each service
RUN cd $workDirRepo/deployments/docker &&\
    python3 generate-config.py

# Build and install binaries, logs, plugins and input directories
RUN make && make install

# Bundle the deploy directory into a tar file for each service
RUN tar -cf provisioning.tar -C ${PROVISIONING_DEPLOY_PREFIX} . &&\
    tar -cf verification.tar -C ${VERIFICATION_DEPLOY_PREFIX} . &&\
    tar -cf vts.tar -C ${VTS_DEPLOY_PREFIX} .


#### Provisioning service
FROM build-base AS provisioning-run

# Set environment variable for input directory for use inside the container
ARG INPUT_FILE_DIR
ENV INPUT_FILE_DIR $INPUT_FILE_DIR

# Renew working directory arg
ARG workDirRepo

WORKDIR /

# Copy over tar ball from build image
COPY --from=common-build $workDirRepo/provisioning.tar /
RUN tar -xf provisioning.tar
EXPOSE 8888
ENTRYPOINT [ "provisioning-service" ]
CMD [ ]


#### Verification service
FROM build-base AS verification-run

# Set environment variable for input directory for use inside the container
ARG INPUT_FILE_DIR
ENV INPUT_FILE_DIR $INPUT_FILE_DIR

# Renew working directory arg
ARG workDirRepo

WORKDIR /

# Copy over tar ball from build image
COPY --from=common-build $workDirRepo/verification.tar /
RUN tar -xf verification.tar && \
    touch /etc/machine-id
EXPOSE 8080
ENTRYPOINT [ "verification-service" ]
CMD [ ]


#### VTS (Veraison Trusted Services)
FROM build-base AS vts-run

# Renew working directory arg
ARG workDirRepo
WORKDIR /

# Copy over tar ball from build image
COPY --from=common-build $workDirRepo/vts.tar /
RUN tar -xf vts.tar
RUN ["/bin/bash", "-c", "/init-kvstores.sh"]
ENTRYPOINT [ "vts-service" ]
CMD [ ]