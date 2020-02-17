#!/usr/bin/env bash

cmd='curl --fail http://localstack:4572'

set -e

${cmd}
while [ $? -ne 0 ]; do
    echo ""
    ${cmd}
done

echo ""
