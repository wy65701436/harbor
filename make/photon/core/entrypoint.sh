#!/bin/sh

set -e

chown -R harbor:harbor /etc/pki/tls/certs /harbor
chmod u+x /harbor/install_cert.sh /harbor/harbor_core

/harbor/install_cert.sh
exec /harbor/harbor_core
