#!/bin/sh

rm -f nanw.key
rm -f nanw.crt
rm -f nanw.csr

openssl req -newkey rsa:2048 -nodes -sha256 -keyout nanw.key -new -out nanw.csr -config req-nanw.conf

openssl req -text -noout -verify -in nanw.csr
