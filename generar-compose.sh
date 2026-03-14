#!/bin/bash
if [ $# -ne 2 ]; then
  echo "Usage: $0 <archivo_salida> <cantidad_clientes>"
  exit 1
fi

cat << EOF > $1
name: tp0
services:
  server:
    container_name: server
    image: server:latest
    entrypoint: python3 /main.py
    environment:
      - PYTHONUNBUFFERED=1
      - LOGGING_LEVEL=DEBUG
    networks:
      - testing_net
    volumes:
      - ./server/config.ini:/config.ini
EOF


for i in $(seq 1 "$2")
do
cat << EOF >> "$1"
  client${i}:
    container_name: client${i}
    image: client:latest
    entrypoint: /client
    environment:
      - CLI_ID=${i}
      - CLI_LOG_LEVEL=DEBUG
    networks:
      - testing_net
    depends_on:
      - server
    volumes:
      - ./client/config.yaml:/config.yaml
EOF
done


cat << EOF >> "$1"
networks:
  testing_net:
    ipam:
      driver: default
      config:
        - subnet: 172.25.125.0/24
EOF