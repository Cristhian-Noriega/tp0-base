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
    networks:
      - testing_net
    volumes:
      - ./server/config.ini:/config.ini
EOF

NOMBRES=(client1 client2 client3 client4 client5)
APELLIDOS=(Gomez Perez Rodriguez Sanchez Fernandez)
DOCUMENTOS=(12345678 87654321 11223344 44332211 55667788)
NACIMIENTOS=(1990-01-01 1991-02-02 1992-03-03 1993-04-04 1994-05-05)
NUMEROS=(5551234 5555678 5559012 5553456 5557890)


for i in $(seq 1 "$2")
do
idx=$((i-1))
cat << EOF >> "$1"
  client${i}:
    container_name: client${i}
    image: client:latest
    entrypoint: /client
    environment:
      - CLI_ID=${i}
      - NOMBRE=${NOMBRES[$idx]}
      - APELLIDO=${APELLIDOS[$idx]}
      - DOCUMENTO=${DOCUMENTOS[$idx]}
      - NACIMIENTO=${NACIMIENTOS[$idx]}
      - NUMERO=${NUMEROS[$idx]}
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