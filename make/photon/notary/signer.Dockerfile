FROM vmware/photon:1.0

RUN tdnf distro-sync -y  || echo \
    && tdnf erase vim -y \
    && tdnf install -y shadow sudo \
    && tdnf clean all \
    && groupadd -r -g 10000 notary \
    && useradd --no-log-init -r -g 10000 -u 10000 notary
COPY ./binary/notary-signer /bin/notary-signer
COPY ./binary/migrate /bin/migrate
COPY ./binary/migrations/ /migrations/
COPY ./signer-start.sh /bin/signer-start.sh

RUN chmod u+x /bin/notary-signer /migrations/migrate.sh /bin/migrate /bin/signer-start.sh
ENV SERVICE_NAME=notary_signer
ENTRYPOINT [ "/bin/signer-start.sh" ]