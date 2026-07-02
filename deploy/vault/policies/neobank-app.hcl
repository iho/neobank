path "secret/data/neobank/*" {
  capabilities = ["read"]
}

path "transit/encrypt/pii" {
  capabilities = ["update"]
}

path "transit/decrypt/pii" {
  capabilities = ["update"]
}

path "transit/hmac/pii-phone" {
  capabilities = ["update"]
}