FROM vmware/photon:1.0

MAINTAINER wangyan@vmware.com

RUN tdnf distro-sync -y \
    && tdnf erase vim -y \
    && tdnf install sudo -y >> /dev/null\
    && tdnf clean all \
    && groupadd -r -g 10000 harbor && useradd --no-log-init -r -g 10000 -u 10000 harbor \
    && mkdir -p /etc/registry

COPY ./make/photon/registry/entrypoint.sh /
COPY ./make/photon/registry/binary/registry /usr/bin
COPY ./make/dev/registry/harbor_registry /harbor/

RUN chmod u+x /harbor/harbor_registry \
    && chmod u+x /usr/bin/registry \
    && chmod u+x /entrypoint.sh \
    && touch /storage/gcresults

HEALTHCHECK CMD curl --fail -s http://127.0.0.1:8080/api/registry/ping

VOLUME ["/var/lib/registry"]
WORKDIR /harbor/
EXPOSE 5000

ENTRYPOINT ["/harbor/entrypoint.sh"]
