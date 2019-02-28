#!/bin/sh

PROXY_BIN="/aws-signing-proxy"
FATE=${FATE:-"standlone"}
FATE_MOUNT=${FATE_MOUNT:-"/tmp/fate"}
FATE_FILE=${FATE_FILE:-"${FATE_MOUNT}/main-terminated"}

if [ "${FATE}" == "shared" ]; then
    echo "Running in shared-fate mode"
    echo "Proxy will terminate when ${FATE_FILE} is present"

    ${PROXY_BIN} ${@} &
    PROXY_PID=$!

    while true; do
        if [[ -f "${FATE_FILE}" ]]; then kill -SIGTERM ${PROXY_PID}; fi
        sleep 2
    done &

    wait ${PROXY_PID}

    if [[ -f "${FATE_FILE}" ]]; then exit 0; fi
else
    echo "Running in standalone mode"
    exec ${PROXY_BIN} ${@}
fi
