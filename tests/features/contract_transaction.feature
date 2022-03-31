@public-tx
Feature: Send contract transactions
  As an external developer
  I want to send multiple contract transactions using transaction-scheduler API

  Background:
    Given I have the following tenants
      | alias   | tenantID        |
      | tenant1 | {{random.uuid}} |
    And I register the following contracts
      | name        | artifacts        | API-KEY            | Tenant               |
      | SimpleToken | SimpleToken.json | {{global.api-key}} | {{tenant1.tenantID}} |
      | Counter     | Counter.json     | {{global.api-key}} | {{tenant1.tenantID}} |
      | Struct      | Struct.json      | {{global.api-key}} | {{tenant1.tenantID}} |
    And I have created the following accounts
      | alias    | ID              | API-KEY            | Tenant               |
      | account1 | {{random.uuid}} | {{global.api-key}} | {{tenant1.tenantID}} |
      | account2 | {{random.uuid}} | {{global.api-key}} | {{tenant1.tenantID}} |
    Then I track the following envelopes
      | ID                  |
      | faucet-{{account1}} |
      | faucet-{{account2}} |
    Given I set the headers
      | Key         | Value                |
      | X-API-KEY   | {{global.api-key}}   |
      | X-TENANT-ID | {{tenant1.tenantID}} |
    When I send "POST" request to "{{global.api}}/transactions/transfer" with json:
      """
      {
        "chain": "{{chain.besu0.Name}}",
        "params": {
          "from": "{{global.nodes.besu[0].fundedPublicKeys[0]}}",
          "to": "{{account1}}",
          "value": "0x16345785D8A0000"
        },
        "labels": {
          "scenario.id": "{{scenarioID}}",
          "id": "faucet-{{account1}}"
        }
      }
      """
    Then the response code should be 202
    When I send "POST" request to "{{global.api}}/transactions/transfer" with json:
      """
      {
        "chain": "{{chain.geth0.Name}}",
        "params": {
          "from": "{{global.nodes.geth[0].fundedPublicKeys[0]}}",
          "to": "{{account2}}",
          "value": "0x16345785D8A0000"
        },
        "labels": {
          "scenario.id": "{{scenarioID}}",
          "id": "faucet-{{account2}}"
        }
      }
      """
    Then the response code should be 202
    Then TxResponse was found in tx-decoded topic "{msgID}"
    And Envelopes should have the following fields
      | Receipt.Status |
      | 1              |
      | 1              |
    Given I register the following alias
      | alias              | value           |
      | besuContractTxID   | {{random.uuid}} |
      | gethContractTxID   | {{random.uuid}} |
      | structContractTxID | {{random.uuid}} |
    Then I track the following envelopes
      | ID                     |
      | {{besuContractTxID}}   |
      | {{gethContractTxID}}   |
      | {{structContractTxID}} |
    When I send "POST" request to "{{global.api}}/transactions/deploy-contract" with json:
      """
      {
        "chain": "{{chain.besu0.Name}}",
        "params": {
          "contractName": "SimpleToken",
          "from": "{{account1}}"
        },
        "labels": {
          "scenario.id": "{{scenarioID}}",
          "id": "{{besuContractTxID}}"
        }
      }
      """
    Then the response code should be 202
    When I send "POST" request to "{{global.api}}/transactions/deploy-contract" with json:
      """
      {
        "chain": "{{chain.geth0.Name}}",
        "params": {
          "contractName": "SimpleToken",
          "from": "{{account2}}"
        },
        "labels": {
          "scenario.id": "{{scenarioID}}",
          "id": "{{gethContractTxID}}"
        }
      }
      """
    Then the response code should be 202
    When I send "POST" request to "{{global.api}}/transactions/deploy-contract" with json:
      """
      {
        "chain": "{{chain.geth0.Name}}",
        "params": {
          "contractName": "Struct",
          "from": "{{account2}}"
        },
        "labels": {
          "scenario.id": "{{scenarioID}}",
          "id": "{{structContractTxID}}"
        }
      }
      """
    Then the response code should be 202
    Then TxResponse was found in tx-decoded topic "{msgID}"
    And I register the following envelope fields
      | id                     | alias              | path                    |
      | {{besuContractTxID}}   | besuContractAddr   | Receipt.ContractAddress |
      | {{gethContractTxID}}   | gethContractAddr   | Receipt.ContractAddress |
      | {{structContractTxID}} | structContractAddr | Receipt.ContractAddress |

  @geth
  Scenario: Send contract transactions
    Given I register the following alias
      | alias             | value           |
      | besuSendTxOneID   | {{random.uuid}} |
      | besuSendTxTwoID   | {{random.uuid}} |
      | besuSendTxThreeID | {{random.uuid}} |
      | gethSendTxOneID   | {{random.uuid}} |
      | gethSendTxTwoID   | {{random.uuid}} |
      | gethSendTxThreeID | {{random.uuid}} |
    Then I track the following envelopes
      | ID                    |
      | {{besuSendTxOneID}}   |
      | {{besuSendTxTwoID}}   |
      | {{besuSendTxThreeID}} |
      | {{gethSendTxOneID}}   |
      | {{gethSendTxTwoID}}   |
      | {{gethSendTxThreeID}} |
    Given I set the headers
      | Key         | Value                |
      | X-API-KEY   | {{global.api-key}}   |
      | X-TENANT-ID | {{tenant1.tenantID}} |
    When I send "POST" request to "{{global.api}}/transactions/send" with json:
      """
      {
        "chain": "{{chain.besu0.Name}}",
        "params": {
          "from": "{{account1}}",
          "to": "{{besuContractAddr}}",
          "contractName":"SimpleToken",
          "methodSignature": "transfer(address,uint256)",
          "args": [
            "0xdbb881a51CD4023E4400CEF3ef73046743f08da3",
            "1"
          ]
        },
        "labels": {
          "scenario.id": "{{scenarioID}}",
          "id": "{{besuSendTxOneID}}"
        }
      }
      """
    Then the response code should be 202
    When I send "POST" request to "{{global.api}}/transactions/send" with json:
      """
      {
        "chain": "{{chain.besu0.Name}}",
        "params": {
          "from": "{{account1}}",
          "to": "{{besuContractAddr}}",
          "contractName":"SimpleToken",
          "methodSignature": "transfer(address,uint256)",
          "args": [
            "0x6009608A02a7A15fd6689D6DaD560C44E9ab61Ff",
            "0x2"
          ]
        },
        "labels": {
          "scenario.id": "{{scenarioID}}",
          "id": "{{besuSendTxTwoID}}"
        }
      }
      """
    Then the response code should be 202
    When I send "POST" request to "{{global.api}}/transactions/send" with json:
      """
      {
        "chain": "{{chain.besu0.Name}}",
        "params": {
          "from": "{{account1}}",
          "to": "{{besuContractAddr}}",
          "contractName":"SimpleToken",
          "methodSignature": "transfer(address,uint256)",
          "gas": 100000,
          "args": [
            "0x6009608A02a7A15fd6689D6DaD560C44E9ab61Ff",
            "0x8"
          ]
        },
        "labels": {
          "scenario.id": "{{scenarioID}}",
          "id": "{{besuSendTxThreeID}}"
        }
      }
      """
    Then the response code should be 202
    When I send "POST" request to "{{global.api}}/transactions/send" with json:
      """
      {
        "chain": "{{chain.geth0.Name}}",
        "params": {
          "from": "{{account2}}",
          "to": "{{gethContractAddr}}",
          "contractName":"SimpleToken",
          "methodSignature": "transfer(address,uint256)",
          "args": [
            "0xdbb881a51CD4023E4400CEF3ef73046743f08da3",
            "0x1"
          ]
        },
        "labels": {
          "scenario.id": "{{scenarioID}}",
          "id": "{{gethSendTxOneID}}"
        }
      }
      """
    Then the response code should be 202
    When I send "POST" request to "{{global.api}}/transactions/send" with json:
      """
      {
        "chain": "{{chain.geth0.Name}}",
        "params": {
          "from": "{{account2}}",
          "to": "{{gethContractAddr}}",
          "contractName":"SimpleToken",
          "methodSignature": "transfer(address,uint256)",
          "args": [
            "0x6009608A02a7A15fd6689D6DaD560C44E9ab61Ff",
            "0x2"
          ]
        },
        "labels": {
          "scenario.id": "{{scenarioID}}",
          "id": "{{gethSendTxTwoID}}"
        }
      }
      """
    Then the response code should be 202
    When I send "POST" request to "{{global.api}}/transactions/send" with json:
      """
      {
        "chain": "{{chain.geth0.Name}}",
        "params": {
          "from": "{{account2}}",
          "to": "{{gethContractAddr}}",
          "contractName":"SimpleToken",
          "methodSignature": "transfer(address,uint256)",
          "gas": 100000,
          "args": [
            "0x6009608A02a7A15fd6689D6DaD560C44E9ab61Ff",
            "2"
          ]
        },
        "labels": {
          "scenario.id": "{{scenarioID}}",
          "id": "{{gethSendTxThreeID}}"
        }
      }
      """
    Then the response code should be 202
    When I send "POST" request to "{{global.api}}/transactions/send" with json:
      """
      {
        "chain": "{{chain.geth0.Name}}",
        "params": {
          "from": "{{account2}}",
          "to": "{{structContractAddr}}",
          "contractName":"Struct",
          "methodSignature": "multipleTransfer((address,(uint256,address)[]))",
          "args": [
              {
                  "token": "0x71b7d704598945e72e7581bac3b070d300dc6eb3",
                  "recipients": [
                      {
                          "amount": 500,
                          "recipient": "0x8015C426A8F5F4673F477b5AaA56756e7193A5A0"
                      }
                  ]
              }
          ]
        },
        "labels": {
          "scenario.id": "{{scenarioID}}"
        }
      }
      """
    Then the response code should be 202
    Then TxResponse was found in tx-decoded topic "{msgID}"
    And Envelopes should have the following fields
      | Receipt.Status | Receipt.Logs[0].Event             | Receipt.Logs[0].DecodedData.from | Receipt.Logs[0].DecodedData.to             | Receipt.Logs[0].DecodedData.value | Receipt.ContractName | Receipt.ContractTag |
      | 1              | Transfer(address,address,uint256) | {{account1}}                     | 0xdbb881a51CD4023E4400CEF3ef73046743f08da3 | 1                                 | SimpleToken          | latest              |
      | 1              | Transfer(address,address,uint256) | {{account1}}                     | 0x6009608A02a7A15fd6689D6DaD560C44E9ab61Ff | 2                                 | SimpleToken          | latest              |
      | 1              | Transfer(address,address,uint256) | {{account1}}                     | 0x6009608A02a7A15fd6689D6DaD560C44E9ab61Ff | 8                                 | SimpleToken          | latest              |
      | 1              | Transfer(address,address,uint256) | {{account2}}                     | 0xdbb881a51CD4023E4400CEF3ef73046743f08da3 | 1                                 | SimpleToken          | latest              |
      | 1              | Transfer(address,address,uint256) | {{account2}}                     | 0x6009608A02a7A15fd6689D6DaD560C44E9ab61Ff | 2                                 | SimpleToken          | latest              |
      | 1              | Transfer(address,address,uint256) | {{account2}}                     | 0x6009608A02a7A15fd6689D6DaD560C44E9ab61Ff | 2                                 | SimpleToken          | latest              |

  @oneTimeKey
  Scenario: Send contract transactions with one-time-key
    Given I register the following alias
      | alias             | value           |
      | counterDeployTxID | {{random.uuid}} |
    Then I track the following envelopes
      | ID                    |
      | {{counterDeployTxID}} |
    Given I set the headers
      | Key         | Value                |
      | X-API-KEY   | {{global.api-key}}   |
      | X-TENANT-ID | {{tenant1.tenantID}} |
    When I send "POST" request to "{{global.api}}/transactions/deploy-contract" with json:
      """
      {
        "chain": "{{chain.besu0.Name}}",
        "params": {
          "oneTimeKey": true,
          "contractName": "Counter"
        },
        "labels": {
          "scenario.id": "{{scenarioID}}",
          "id": "{{counterDeployTxID}}"
        }
      }
      """
    Then the response code should be 202
    Then TxResponse was found in tx-decoded topic "{msgID}"
    And I register the following envelope fields
      | id                    | alias               | path                    |
      | {{counterDeployTxID}} | counterContractAddr | Receipt.ContractAddress |
    Given I register the following alias
      | alias       | value           |
      | sendOTKTxID | {{random.uuid}} |
    Then I track the following envelopes
      | ID              |
      | {{sendOTKTxID}} |
    When I send "POST" request to "{{global.api}}/transactions/send" with json:
      """
      {
        "chain": "{{chain.besu0.Name}}",
        "params": {
          "oneTimeKey": true,
          "to": "{{counterContractAddr}}",
          "contractName": "Counter",
          "methodSignature": "increment(uint256)",
          "args": [
            1
          ]
        },
        "labels": {
          "scenario.id": "{{scenarioID}}",
          "id": "{{sendOTKTxID}}"
        }
      }
      """
    Then the response code should be 202
    Then I register the following response fields
      | alias      | path         |
      | jobOTKUUID | jobs[0].uuid |
    Then TxResponse was found in tx-decoded topic "{msgID}"
    And Envelopes should have the following fields
      | Receipt.Status | Receipt.Logs[0].Event        | Receipt.Logs[0].DecodedData.from |
      | 1              | Incremented(address,uint256) | ~                                |
    When I send "GET" request to "{{global.api}}/jobs/{{jobOTKUUID}}"
    Then the response code should be 200
    And Response should have the following fields
      | status | logs[0].status | logs[1].status | logs[2].status | logs[3].status |
      | MINED  | CREATED        | STARTED        | PENDING        | MINED          |

  Scenario: Fail to send contract transactions with invalid args
    Given I register the following alias
      | alias             | value           |
      | counterDeployTxID | {{random.uuid}} |
    Then I track the following envelopes
      | ID                    |
      | {{counterDeployTxID}} |
    Given I set the headers
      | Key         | Value                |
      | X-API-KEY   | {{global.api-key}}   |
      | X-TENANT-ID | {{tenant1.tenantID}} |
    When I send "POST" request to "{{global.api}}/transactions/deploy-contract" with json:
      """
      {
        "chain": "{{chain.besu0.Name}}",
        "params": {
          "from": "{{account1}}",
          "contractName": "Counter"
        },
        "labels": {
          "scenario.id": "{{scenarioID}}",
          "id": "{{counterDeployTxID}}"
        }
      }
      """
    Then the response code should be 202
    Then TxResponse was found in tx-decoded topic "{msgID}"
    And I register the following envelope fields
      | id                    | alias               | path                    |
      | {{counterDeployTxID}} | counterContractAddr | Receipt.ContractAddress |
    Given I register the following alias
      | alias    | value           |
      | sendTxID | {{random.uuid}} |
    Then I track the following envelopes
      | ID           |
      | {{sendTxID}} |
    When I send "POST" request to "{{global.api}}/transactions/send" with json:
      """
      {
        "chain": "{{chain.besu0.Name}}",
        "params": {
          "from": "{{account1}}",
          "to": "{{counterContractAddr}}",
          "contractName": "Counter",
          "methodSignature": "increment(uint256)",
          "args": [
            "notANumber"
          ]
        },
        "labels": {
          "scenario.id": "{{scenarioID}}",
          "id": "{{sendTxID}}"
        }
      }
      """
    Then the response code should be 422
    And Response should have the following fields
      | code   | message |
      | 271360 | ~       |
