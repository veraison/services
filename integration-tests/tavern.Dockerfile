# go version for container image
ARG GoVersion=1.18

# Arguments for psa demo input directories
# TODO create version associated with the tag by adding a release
ARG COCLI_TEMPLATES=/go/pkg/mod/github.com/veraison/corim@v0.0.0-20221125105155-c2835023f15e/cocli
ARG EVCLI_TEMPLATES=/go/pkg/mod/github.com/veraison/evcli@v0.0.0-20221212172836-49c7b2bdcf38/misc

FROM golang:$GoVersion AS tavern-integration-tests
WORKDIR /

ARG COCLI_TEMPLATES
ARG EVCLI_TEMPLATES

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
    && apt-get clean \
    && apt-get autoremove --assume-yes \
    && rm -rf /var/lib/apt/lists/* /var/tmp/* /tmp/*

RUN go install github.com/veraison/corim/cocli@demo-psa-1.0.1 &&\
    go install github.com/veraison/evcli@demo-psa-1.0.1 &&\
    pip3 install tavern &&\
    pip3 install python-jose


COPY integration-tests/ /integration-tests/

WORKDIR /integration-tests/extra
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

WORKDIR /integration-tests/verification
RUN wget --progress=dot:giga https://raw.githubusercontent.com/veraison/services/main/vts/cmd/vts-service/skey.jwk 

WORKDIR /
CMD PYTHONPATH=$PYTHONPATH:integration-tests py.test integration-tests/ -vv
