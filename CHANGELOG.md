# CHANGELOG

All notable changes to this project will be documented in this file.

## v1.2.1 (2019-20-13)

### 🛠 Bug fixes

* Upgrade retry policy when getting `NotFoundError` on JSON-RPC request. In particular it allows the transaction listener to effectively handle Infura endpoint that have sync discrepancies.

## v1.2.0 (2019-20-13)

### 🆕 Features

* Add new flag and environment variable `REDIS_EXPIRATION` to configure Redis entry expiration. It is useful for `tx-nonce` and `tx-sender` workers to expire keys and force a nonce recalibration from chain after inactivity of a sender
* Add new flag and environment variable `NONCE_CHECKER_MAX_RECOVERY` to configure max number of nonce recoveries to perform on a given envelope on `tx-sender`
* Nonce checker on `tx-sender` ignores envelopes with metadata entry `tx.mode` set to `raw`
* Add new flag and environment variable `DISABLE_EXTERNAL_TX` in the tx-listener to filter transactions not sent through Orchestrate

### 🛠 Bug fixes

* Fix connection issue when trying to connect to some Infura endpoints
* On `envelopestore`, when storing a 2nd envelope with same `tx_hash` and `chain_id` but a distinct `metadata.id` overwrites the first one
* Fix exposition of Swagger-UI in Docker images
* Fix crafting transactions with other types from uint256 and int256

### ⚠ BREAKING CHANGES
* **config** Rename `mock` options to `in-memory` for the `NONCE_MANAGER_TYPE` of the `tx-nonce` and `tx-sender` 


## 1.0.2 (2019-12-18)

### 🛠 Bug fixes
* Fix issue when registering a contract with no methods and/or no events


## v1.1.0 (2019-12-10)

### 🆕 Features

* Add a new server for APIs services exposing a REST endpoint that will redirect queries to the gRPC endpoint and a swagger UI + documentation
* Add the new flag `KAFKA_CONSUMER_MAX_WAIT_TIME` to configure the maximum waiting time to consume message if messages do not exceed the size`Consumer.Fetch.Min.Byte` (default=20ms)

### 🛠 Bug fixes
* Clean logs: downgrade `OK` and `NotFound` logs to debug level in grpc server and have logger handler in debug level for tx-listener and tx-decoder
* e2e: use `CUCUMBER_STEPS_TIMEOUT` to put a timeout in cucumber steps before failing
* Makefile: add bootstrap stage to wait quorum, geth and kafka to start

### ⚠ BREAKING CHANGES
* **config** grpc & metrics server have been split. Default port for
  * grpc server remains 8080
  * newly rest server is 8081 
  * metrics server has changed and is now 8082
* **config** Rename `KAFKA_SASL_ENABLE` to `KAFKA_SASL_ENABLED`


## 1.0.1 (2019-12-10)

### 🛠 Bug fixes
* Fix issue when registering a contract with methods/events having the same name and different signatures


## 1.0.0 (2019-11-07)

This is the first stable release of Orchestrate.

### 🆕 Features

* **Transaction Management**, automatically manage transactions lifecycle
    * Transaction crafting: craft transactions and deploy contract based on the bytecode in the Contract Registry
    * Faucet: get accounts credited with Ether
    * Gas management: gas price and limit can be provided or estimated by the Ethereum client
    * Transaction nonce management: set transactions nonce automatically, avoid Nonce too high scenario
    * Transaction and Event listening: process mined transactions’ receipts *at least once* and attempt to decode them with ABIs in the Contract Registry
    * External transaction signing & Key storage: store keys in HashiCorp Vault
    * External Private Transaction Signature on Besu-Orion and Quorum-Tessera.
* **Contract-Registry**: Dynamically register Smart Contract Artifacts (ABIs, bytecode & deployedBytecode) referenced by a (name, tag). Currently available in Postgres
* Connect to multiple blockchain networks simultaneously (multiple jsonRPC endpoints)
* Connect to EVM based blockchain (public/private, [Hyperledger Besu](https://github.com/hyperledger/besu), Go-Ethereum, Parity, Quorum+Tessera/Constellation)
* Add Logging using Logrus, Prometheus metrics exporter and OpenTracing exporter capabilities

### ⚠ BREAKING CHANGES

This is the list of breaking change with the Beta release. 

* **config** Rename `JAEGER_DISABLED` to `JAEGER_ENABLED`
* **config** Rename `GRPC_TARGET_CONTRACT_REGISTRY` to `CONTRACT_REGISTRY_URL`
* **config** Rename `GRPC_TARGET_ENVELOPE_STORE` to `ENVELOPE_STORE_URL`
* **config** Rename `KAFKA_ADDRESS` to `KAFKA_URL`
* **config** Rename `REDIS_ADDRESS` to `REDIS_URL`
* **config** Rename `TESSERA_ENDPOINTS` to `TESSERA_URL`
* **config** Rename `mock` options to `in-memory`
* **config** Remove `--engine-slot`
* **config** Rename `FAUCET` to `FAUCET_TYPE` & remove `--faucet`
* **config** Rename `VAULT_TOKEN_FILEPATH` to `VAULT_TOKEN_FILE`
* **config** Rename `KAFKA_TOPIC_TX_XXX` to `TOPIC_TX_XXXX`
* **config** Rename `KAFKA_TLS_CLIENTCERTFILEPATH` to `KAFKA_TLS_CLIENT_CERT_FILE`
* **config** Rename `KAFKA_TLS_CLIENTKEYFILEPATH` to `KAFKA_TLS_CLIENT_KEY_FILE`
* **config** Rename `KAFKA_TLS_CACERTFILEPATH` to `KAFKA_TLS_CA_CERT_FILE`
* **config** Rename `KAFKA_TLS_INSECURESKIPVERIFY` to `KAFKA_TLS_INSECURE_SKIP_VERIFY`
* **config** Rename `KAFKA_TLS_ENABLE` to `KAFKA_TLS_ENABLED`
* **config** Rename `NONCE_MANAGER` to `NONCE_MANAGER_TYPE`
* **config** Rename `CONTRACT_REGISTRY` to `CONTRACT_REGISTRY_TYPE`
* **config** Rename `ENVELOPE_STORE` to `ENVELOPE_STORE_TYPE`
* **config** Rename Envelope-store option `pg` to `postgres`
