ARG harbor_base_image_version
ARG harbor_base_namespace
FROM ${harbor_base_namespace}/harbor-core-base:${harbor_base_image_version}

HEALTHCHECK CMD curl --fail -s http://localhost:8080/api/v2.0/ping || curl -k --fail -s https://localhost:8443/api/v2.0/ping || exit 1
COPY ./make/photon/common/install_cert.sh \
     ./make/photon/core/entrypoint.sh \
     ./make/photon/core/harbor_core \
     ./src/core/views \
     ./make/migrations \
     ./icons \
     ./make/photon/prepare/versions /harbor/

RUN chown harbor:harbor /harbor/entrypoint.sh && chmod u+x /harbor/entrypoint.sh

WORKDIR /harbor/
USER harbor
ENTRYPOINT ["/harbor/entrypoint.sh"]
