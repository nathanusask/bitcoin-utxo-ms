# bitcoin-utxo-ms

To simply build and run

```shell
go build

#dev-testnet
set -o allexport; source dev-testnet.env; set +o allexport
./bitcoin-utxo-ms > btc-utxo-ms-dev-testnet.log 2>&1 &

#dev-mainnet
set -o allexport; source dev-mainnet.env; set +o allexport
./bitcoin-utxo-ms > btc-utxo-ms-dev-mainnet.log 2>&1 &

#prod-testnet
set -o allexport; source prod-testnet.env; set +o allexport
./bitcoin-utxo-ms > btc-utxo-ms-prod-testnet.log 2>&1 &

#prod-mainnet
set -o allexport; source prod-mainnet.env; set +o allexport
./bitcoin-utxo-ms > btc-utxo-ms-prod-mainnet.log 2>&1 &

```

To deploy with docker-compose
Run `sudo docker-compose up -d`