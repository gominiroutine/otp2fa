#!/bin/bash
# shellcheck disable=SC2164
cd "$(dirname "$0")"

docker stop otp2fa
docker rm otp2fa
docker rmi manhavn/otp2fa:v0.0.1
docker run -d --name otp2fa \
  -v ${PWD}/env:/app/env \
  -v ${PWD}/database:/app/database \
  -v ${PWD}/qrcode:/app/qrcode \
  -v ${PWD}/new-qrcode:/app/new-qrcode \
  -it manhavn/otp2fa:v0.0.1
