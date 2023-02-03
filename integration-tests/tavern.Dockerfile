# go version for container image
ARG GoVersion=1.18

FROM golang:$GoVersion AS tavern-integration-tests
WORKDIR /

ARG COCLI_TEMPLATES
ARG EVCLI_TEMPLATES
ARG DIAG_FILES

# Set environment variables for psa demo input directories
ENV COCLI_TEMPLATES $COCLI_TEMPLATES
ENV EVCLI_TEMPLATES $EVCLI_TEMPLATES
ENV DIAG_FILES $DIAG_FILES

ARG PROVISIONING_CONTAINER_NAME
ARG VERIFICATION_CONTAINER_NAME
ARG VTS_CONTAINER_NAME
ARG VTS_PROVISIONING_NETWORK_ALIAS
ARG VTS_VERIFICATION_NETWORK_ALIAS

# Set envrionment variables for host name and netowrk aliases
ENV PROVISIONING_CONTAINER_NAME $PROVISIONING_CONTAINER_NAME
ENV VERIFICATION_CONTAINER_NAME $VERIFICATION_CONTAINER_NAME
ENV VTS_CONTAINER_NAME $VTS_CONTAINER_NAME
ENV VTS_PROVISIONING_NETWORK_ALIAS $VTS_PROVISIONING_NETWORK_ALIAS
ENV VTS_VERIFICATION_NETWORK_ALIAS $VTS_VERIFICATION_NETWORK_ALIAS

RUN apt-get update \
    && DEBIAN_FRONTEND=noninteractive apt-get install \
        --assume-yes \
        --no-install-recommends \
        apt-transport-https \
        apt-utils \
        python3-pip \
        ruby-full \
        xxd \
    && apt-get clean \
    && apt-get autoremove --assume-yes \
    && rm -rf /var/lib/apt/lists/* /var/tmp/* /tmp/*

#  Install tools
RUN go install github.com/veraison/corim/cocli@demo-psa-1.0.1 &&\
    go install github.com/veraison/evcli@demo-psa-1.0.1 &&\
    go install github.com/thomas-fossati/go-cose-cli@latest &&\
    go install github.com/thomas-fossati/go-psa@latest &&\
    pip3 install tavern &&\
    pip3 install python-jose &&\
    pip install cbor-json &&\
    gem install cbor-diag

# Make test vector directories
RUN mkdir -p /test-vectors/provisioning/cbor &&\
    mkdir -p /test-vectors/provisioning/json &&\
    mkdir -p /test-vectors/provisioning/keys &&\
    mkdir -p /test-vectors/verification/cbor &&\
    mkdir -p /test-vectors/verification/json &&\
    mkdir -p /test-vectors/verification/keys

# Copy over important keys and base templates
RUN cp $COCLI_TEMPLATES/data/keys/ec-p256.jwk /test-vectors/verification/keys &&\
    cp $EVCLI_TEMPLATES/ec256.json /test-vectors/verification/keys &&\
    cp $EVCLI_TEMPLATES/psa-claims-profile-2-integ.json /test-vectors/verification/json &&\
    cp $EVCLI_TEMPLATES/psa-claims-profile-2-integ-without-nonce.json test-vectors/verification/json

WORKDIR /test-vectors/verification/keys
RUN wget --progress=dot:giga https://raw.githubusercontent.com/veraison/services/main/vts/cmd/vts-service/skey.jwk 

WORKDIR /
CMD PYTHONPATH=$PYTHONPATH:integration-tests py.test integration-tests/ -vv
