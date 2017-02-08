#!/bin/bash

VAR_FILE=${VAR_FILE:-/tmp/var.yaml}
if [ ! -f ${VAR_FILE} ]; then
    touch ${VAR_FILE}
fi

./main consume --uri ${RABBITMQ_URL} --var ${VAR_FILE}
