#!/bin/bash 

MENSAJE_A_ENVIAR="hola"

MENSAJE_RECIBIDO=$(echo "$MENSAJE_A_ENVIAR" | docker run --rm -i --network tp0_testing_net busybox nc server 12345)

if [ "$MENSAJE_A_ENVIAR" == "$MENSAJE_RECIBIDO" ]; then
    echo "action: test_echo_server | result: success"
else
    echo "action: test_echo_server | result: fail"
fi


