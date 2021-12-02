FROM golang:1.17-alpine

WORKDIR /app

COPY . .

RUN go build -o /btc-utxo-mx

EXPOSE 18088 19088 28088 29088

CMD [ "/btc-utxo-mx" ]