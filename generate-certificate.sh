#!/bin/ash
FILE_CERT_NAME=localcert
if [ -f "certificates/$FILE_CERT_NAME.crt" ] && [ -f "$FILE_CERT_NAME.key" ] ; then
    echo "Cert and Key already exist"
else 
    echo "Cert and Key does not exist, trying to create new ones..."
    apk update && apk add openssl && rm -rf /var/cache/apk/*
    openssl req -new -subj "/C=US/ST=California/CN=localhost" \
        -newkey rsa:2048 -nodes -keyout "$FILE_CERT_NAME.key" -out "$FILE_CERT_NAME.csr"
    openssl x509 -req -days 365 -in "$FILE_CERT_NAME.csr" -signkey "$FILE_CERT_NAME.key" -out "$FILE_CERT_NAME.crt" -extfile "self-signed-cert.ext"
fi