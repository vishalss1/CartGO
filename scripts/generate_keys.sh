#!/bin/bash

# Simple script to generate RS256 keys for CartGO

KEYS_DIR="$(dirname "$0")/../keys"
mkdir -p "$KEYS_DIR"

PRIVATE_KEY="$KEYS_DIR/jwt_private.pem"
PUBLIC_KEY="$KEYS_DIR/jwt_public.pem"

echo "Generating new RSA key pair in $KEYS_DIR..."

# Generate standard 2048-bit RSA private key
openssl genrsa -out "$PRIVATE_KEY" 2048

# Extract the public key
openssl rsa -in "$PRIVATE_KEY" -pubout -out "$PUBLIC_KEY"

echo "Done! Keys generated successfully."
echo "Private Key: $PRIVATE_KEY"
echo "Public Key: $PUBLIC_KEY"
