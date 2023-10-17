#!/bin/sh
export SSL_CERT_DIR=/mnt/onboard/.kobo/certificates

/mnt/onboard/.adds/kavita-sync/kavita-sync --config=/mnt/onboard/example-config.yaml >/mnt/onboard/kavita-sync.log 2>&1