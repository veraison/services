# Args needs to be renewed for each build stage that uses these args
ARG workDir=/app
ARG workDirRepo=/app/services
ARG GoVersion=1.18

# Arguments for psa demo input directories
ARG COCLI_TEMPLATES=/go/pkg/mod/github.com/veraison/corim@v0.0.0-20221125105155-c2835023f15e/cocli
ARG EVCLI_TEMPLATES=/go/pkg/mod/github.com/veraison/evcli@v0.0.0-20221212172836-49c7b2bdcf38/misc

FROM golang:$GoVersion AS tavern-integration-tests
ARG COCLI_TEMPLATES
ARG EVCLI_TEMPLATES

RUN apt-get update \
    && DEBIAN_FRONTEND=noninteractive apt-get install \
        --assume-yes \
        --no-install-recommends \
        apt-transport-https \
        apt-utils \
        python3-pip \
    && apt-get clean \
    && apt-get autoremove --assume-yes \
    && rm -rf /var/lib/apt/lists/* /var/tmp/* /tmp/*

RUN go install github.com/veraison/corim/cocli@demo-psa-1.0.1 &&\
    go install github.com/veraison/evcli@demo-psa-1.0.1 &&\
    pip3 install tavern &&\
    pip3 install cryptography

COPY integration-tests/ /integration-tests/

RUN cocli comid create --template=$COCLI_TEMPLATES/data/templates/comid-psa-integ-iakpub.json &&\
    cocli comid create --template=$COCLI_TEMPLATES/data/templates/comid-psa-refval.json &&\ 
    cocli corim create --template=$COCLI_TEMPLATES/data/templates/corim-full.json --comid=comid-psa-integ-iakpub.cbor --comid=comid-psa-refval.cbor &&\
    mkdir /integration-tests/provisioning/ &&\
    mv corim-full.cbor /integration-tests/provisioning/

RUN evcli psa create -c $EVCLI_TEMPLATES/psa-claims-profile-2-integ.json -k $COCLI_TEMPLATES/data/keys/ec-p256.jwk --token=psa-evidence.cbor &&\
    mkdir /integration-tests/verification/ &&\
    mv psa-evidence.cbor /integration-tests/verification/ &&\
    cp $COCLI_TEMPLATES/data/keys/ec-p256.jwk /integration-tests/verification/ &&\
    cp $EVCLI_TEMPLATES/psa-claims-profile-2-integ.json /integration-tests/verification/ &&\
    cp $EVCLI_TEMPLATES/psa-claims-profile-2-integ-without-nonce.json /integration-tests/verification/

WORKDIR /

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
        iputils-ping \
        jq \
        tree \
        protobuf-compiler \
        python3 \
        python3-pip \
        libprotobuf-dev \
        libsqlite3-dev \
        net-tools \
        sqlite3 \
        vim \
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
    go install github.com/veraison/corim/cocli@demo-psa-1.0.1 &&\
    go install github.com/veraison/evcli@demo-psa-1.0.1


FROM build-base as common-build
# Renew build arguments
ARG workDirRepo
ARG PROVISIONING_DEPLOY_PREFIX
ARG VERIFICATION_DEPLOY_PREFIX
ARG VTS_DEPLOY_PREFIX
ARG VTS_PROVISIONING_LOCAL_IP_ADDRESS_PORT
ARG VTS_VERIFICATION_LOCAL_IP_ADDRESS_PORT
ARG BIN_DIR
ARG LOG_DIR
ARG PLUGIN_DIR
ARG INPUT_FILE_DIR
ARG GIT_CLONE_BRANCH
ARG PROVISIONING_LISTENING_ADDRESS_AND_PORT
ARG VERIFICATION_LISTENING_ADDRESS_AND_PORT
ARG JWT_SIGNING_ALGORITHM
ARG JWT_KEY_FILE_PATH

# Set envrionment variables for deploy directories and communication ports
ENV PROVISIONING_DEPLOY_PREFIX $workDirRepo$PROVISIONING_DEPLOY_PREFIX
ENV VERIFICATION_DEPLOY_PREFIX $workDirRepo$VERIFICATION_DEPLOY_PREFIX
ENV VTS_DEPLOY_PREFIX $workDirRepo$VTS_DEPLOY_PREFIX
ENV VTS_PROVISIONING_LOCAL_IP_ADDRESS_PORT $VTS_PROVISIONING_LOCAL_IP_ADDRESS_PORT
ENV VTS_VERIFICATION_LOCAL_IP_ADDRESS_PORT $VTS_VERIFICATION_LOCAL_IP_ADDRESS_PORT

# Set environment variables for binaries, logs, plugins and input directories
ENV BIN_DIR $BIN_DIR
ENV LOG_DIR $LOG_DIR
ENV PLUGIN_DIR $PLUGIN_DIR
ENV INPUT_FILE_DIR $INPUT_FILE_DIR

# Set environment variable for branch/tag to git clone from
ENV GIT_CLONE_BRANCH $GIT_CLONE_BRANCH

# Set environment variable for provisioning and verification listening address and port
ENV PROVISIONING_LISTENING_ADDRESS_AND_PORT $PROVISIONING_LISTENING_ADDRESS_AND_PORT
ENV VERIFICATION_LISTENING_ADDRESS_AND_PORT $VERIFICATION_LISTENING_ADDRESS_AND_PORT

# Set environment variable for JWT signing algorithm and key location
ENV JWT_SIGNING_ALGORITHM $JWT_SIGNING_ALGORITHM
ENV JWT_KEY_FILE_PATH $JWT_KEY_FILE_PATH

# Clone main branch from remote (at the demo tag)
RUN git clone --depth 1 --branch ${GIT_CLONE_BRANCH} https://github.com/veraison/services.git

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

# Create input files for psa provisioning part of the demo
ARG COCLI_TEMPLATES
RUN cocli comid create --template=$COCLI_TEMPLATES/data/templates/comid-psa-integ-iakpub.json &&\
    cocli comid create --template=$COCLI_TEMPLATES/data/templates/comid-psa-refval.json &&\ 
    cocli corim create --template=$COCLI_TEMPLATES/data/templates/corim-full.json --comid=comid-psa-integ-iakpub.cbor --comid=comid-psa-refval.cbor &&\
    mv corim-full.cbor $INPUT_FILE_DIR/

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

# Create input files for psa verification part of the demo
ARG COCLI_TEMPLATES
ARG EVCLI_TEMPLATES
RUN evcli psa create -c $EVCLI_TEMPLATES/psa-claims-profile-2-integ.json -k $COCLI_TEMPLATES/data/keys/ec-p256.jwk --token=psa-evidence.cbor &&\
    mv psa-evidence.cbor $INPUT_FILE_DIR/ &&\
    cp $COCLI_TEMPLATES/data/keys/ec-p256.jwk $INPUT_FILE_DIR/ &&\
    cp $EVCLI_TEMPLATES/psa-claims-profile-2-integ-without-nonce.json $INPUT_FILE_DIR/


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
