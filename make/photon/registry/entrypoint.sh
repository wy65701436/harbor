#!/bin/sh

set -e

chown -R harbor:harbor /etc/pki/tls/certs /home/harbor/install_cert.sh /usr/bin/registry_DO_NOT_USE_GC \
    && chmod u+x /home/harbor/install_cert.sh /usr/bin/registry_DO_NOT_USE_GC

/home/harbor/install_cert.sh

exec /usr/bin/registry_DO_NOT_USE_GC serve /etc/registry/config.yml
