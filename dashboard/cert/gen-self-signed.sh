#!/bin/sh

rm -f self-signed-root.key
rm -f self-signed-root.crt
rm -f self-signed-root.srl

# Generate self-signed root CA
openssl req -config req-self-signed-root.conf -newkey rsa:2048 -x509 -days 3650 -nodes -sha256 -out self-signed-root.crt -keyout self-signed-root.key

#openssl x509 -text -noout -purpose -ocsp_uri -in self-signed-root.crt

# Verify
#openssl rsa -check -noout -in self-signed-root.key
#openssl x509 -noout -modulus -in self-signed-root.crt | openssl md5
#openssl rsa -noout -modulus -in self-signed-root.key | openssl md5

# Sign nanw
openssl x509 -req -days 3650 -sha256 -CA self-signed-root.crt -CAkey self-signed-root.key -CAcreateserial -in nanw.csr -extfile v3_ext.txt -out nanw.crt 

# Verify
#openssl rsa -check -noout -in nanw.key
#openssl x509 -noout -modulus -in nanw.crt | openssl md5
#openssl rsa -noout -modulus -in nanw.key | openssl md5

#openssl verify -verbose -CAfile self-signed-root.crt nanw.crt

openssl x509 -text -noout -purpose -ocsp_uri -in nanw.crt
