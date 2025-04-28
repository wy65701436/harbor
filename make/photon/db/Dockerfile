ARG harbor_base_image_version
ARG harbor_base_namespace
FROM ${harbor_base_namespace}/harbor-db-base:${harbor_base_image_version}

VOLUME /var/lib/postgresql/data

COPY ./make/photon/db/docker-entrypoint.sh \
     ./make/photon/db/initdb.sh \
     ./make/photon/db/upgrade.sh \
     ./make/photon/db/docker-healthcheck.sh \
     ./make/photon/db/initial-registry.sql \
     /

RUN chown postgres:postgres /docker-entrypoint.sh && chmod u+x /docker-entrypoint.sh

ENTRYPOINT ["/docker-entrypoint.sh", "", "15"]
HEALTHCHECK CMD ["/docker-healthcheck.sh"]

USER postgres
