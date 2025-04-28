ARG harbor_base_image_version
ARG harbor_base_namespace
FROM ${harbor_base_namespace}/harbor-registryctl-base:${harbor_base_image_version}

COPY ./make/photon/common/install_cert.sh /home/harbor
COPY ./make/photon/registry/binary/registry /usr/bin/registry_DO_NOT_USE_GC
COPY ./make/photon/registryctl/start.sh /home/harbor
COPY ./make/photon/registryctl/harbor_registryctl /home/harbor

RUN chown harbor:harbor /home/harbor/start.sh && chmod u+x /home/harbor/start.sh

HEALTHCHECK CMD curl --fail -s http://localhost:8080/api/health || curl -sk --fail --key /etc/harbor/ssl/registryctl.key --cert /etc/harbor/ssl/registryctl.crt https://localhost:8443/api/health || exit 1

VOLUME ["/var/lib/registry"]

ENTRYPOINT ["/home/harbor/start.sh"]

USER harbor
