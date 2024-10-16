#!/bin/bash
# shellcheck disable=SC2164
# shellcheck disable=SC2016
cd "$(dirname "$0")"

sh format.sh; sh vendor.sh

DI_PROJECT=manhavn
if [ "$2" ]; then
  DI_PROJECT=$2
fi
DI_PACKAGE=otp2fa
DI_VERSION=v0.0.1

TAG_NAME=$DI_PROJECT/$DI_PACKAGE:$DI_VERSION

#docker buildx build --platform linux/amd64 -t "$TAG_NAME" --load .
docker buildx build --platform linux/amd64,linux/arm64 -t "$TAG_NAME" --push .
#docker stop buildx_buildkit_container-builder0

echo ' docker run -d --name otp2fa -v ${PWD}/env:/app/env -v ${PWD}/database:/app/database -v ${PWD}/qrcode:/app/qrcode -v ${PWD}/new-qrcode:/app/new-qrcode -it '"$TAG_NAME"
echo ' docker exec otp2fa create --issuer="test.com" --account="hello@account.com" --title="Test Title"'
echo ' docker exec otp2fa load --database="totp.db" --qrcode="test-com-Test-Title-hello-account-com.png"'
echo ' docker exec -it otp2fa update --database="totp.db"'
echo ' docker exec -it otp2fa genqr --database="totp.db" --output="new-qrcode"'
echo ' docker exec -it otp2fa otp --database="totp.db"'
echo ' docker exec -it otp2fa remove --database="totp.db"'

# git fetch --prune; git reset --hard origin/main;
