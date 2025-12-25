# TCP

## IPv4

ss -tln | grep 55080
ss -tln | grep 55443

## IPv6

ss -tln | grep 55080 | grep "\[::\]"
ss -tln | grep 55443 | grep "\[::\]"

# UDP

## IPv4

ss -uln | grep 55080
ss -uln | grep 55443

## IPv6

ss -uln | grep 55080 | grep "\[::\]"
ss -uln | grep 55443 | grep "\[::\]"
