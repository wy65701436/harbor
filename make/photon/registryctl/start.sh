#!/bin/sh

set -e

chown -R harbor:harbor /etc/pki/tls/certs /home/harbor/harbor_registryctl /usr/bin/registry_DO_NOT_USE_GC /home/harbor/install_cert.sh \
    && chmod u+x /home/harbor/harbor_registryctl /usr/bin/registry_DO_NOT_USE_GC /home/harbor/install_cert.sh

/home/harbor/install_cert.sh

exec /home/harbor/harbor_registryctl -c /etc/registryctl/config.yml
