@nonce
Feature: Nonce manager
  As an external developer
  I want transaction with empty nonce to be calibrated, sent to blockchain and mined

  Background:
    Given I have the following tenants
      | alias   | tenantID        |
      | tenant1 | {{random.uuid}} |
    And I register the following contracts
      | name        | artifacts        | API-KEY            | Tenant               |
      | SimpleToken | SimpleToken.json | {{global.api-key}} | {{tenant1.tenantID}} |
    And I have created the following accounts
      | alias    | ID              | API-KEY            | Tenant               |
      | account1 | {{random.uuid}} | {{global.api-key}} | {{tenant1.tenantID}} |
      | account2 | {{random.uuid}} | {{global.api-key}} | {{tenant1.tenantID}} |
      | account3 | {{random.uuid}} | {{global.api-key}} | {{tenant1.tenantID}} |
    Then I track the following envelopes
      | ID                  |
      | faucet-{{account1}} |
      | faucet-{{account2}} |
      | faucet-{{account3}} |
#      | faucet-geth-{{account1}} |
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
    "chain": "{{chain.besu0.Name}}",
    "params": {
      "from": "{{global.nodes.besu[0].fundedPublicKeys[0]}}",
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
    When I send "POST" request to "{{global.api}}/transactions/transfer" with json:
  """
{
    "chain": "{{chain.besu0.Name}}",
    "params": {
      "from": "{{global.nodes.besu[0].fundedPublicKeys[0]}}",
      "to": "{{account3}}",
      "value": "0x16345785D8A0000"
    },
    "labels": {
    	"scenario.id": "{{scenarioID}}",
    	"id": "faucet-{{account3}}"
    }
}
      """
    Then the response code should be 202
    Then Envelopes should be in topic "tx.decoded"
#    When I send "POST" request to "{{global.api}}/transactions/transfer" with json:
#      """
#      {
#        "chain": "{{chain.geth0.Name}}",
#        "params": {
#          "from": "{{global.nodes.geth[0].fundedPublicKeys[0]}}",
#          "to": "{{account1}}",
#          "value": "0x56BC75E2D63100000"
#        },
#        "labels": {
#          "scenario.id": "{{scenarioID}}",
#          "id": "faucet-geth-{{account1}}"
#        }
#      }
#      """
#    Then the response code should be 202
#    Then Envelopes should be in topic "tx.decoded"

  Scenario: Nonce recalibrating
    Given I register the following alias
      | alias | value              |
      | to1   | {{random.account}} |
      | to2   | {{random.account}} |
      | to3   | {{random.account}} |
    Then I set the headers
      | Key         | Value                |
      | X-API-KEY   | {{global.api-key}}   |
      | X-TENANT-ID | {{tenant1.tenantID}} |
    When I send "POST" request to "{{global.api}}/schedules" with json:
      """
{}
      """
    Then the response code should be 200
    Then I register the following response fields
      | alias           | path |
      | scheduleOneUUID | uuid |
    When I send "POST" request to "{{global.api}}/schedules" with json:
      """
{}
      """
    Then the response code should be 200
    Then I register the following response fields
      | alias           | path |
      | scheduleTwoUUID | uuid |
    When I send "POST" request to "{{global.api}}/schedules" with json:
      """
{}
      """
    Then the response code should be 200
    Then I register the following response fields
      | alias             | path |
      | scheduleThreeUUID | uuid |
    When I send "POST" request to "{{global.api}}/jobs" with json:
  """
{
    "scheduleUUID": "{{scheduleOneUUID}}",
	"chainUUID": "{{chain.besu0.UUID}}",
    "type": "eth://ethereum/transaction",
    "transaction": {
        "from": "{{account1}}",
        "to": "{{to1}}",
        "value": "0x222E0",
        "gas": 21000
    },
    "labels": {
    	"scenario.id": "{{scenarioID}}"
    }
}
      """
    Then the response code should be 200
    Then I register the following response fields
      | alias        | path |
      | txOneJobUUID | uuid |
    When I send "POST" request to "{{global.api}}/jobs" with json:
  """
{
    "scheduleUUID": "{{scheduleTwoUUID}}",
	"chainUUID": "{{chain.besu0.UUID}}",
    "type": "eth://ethereum/transaction",
    "transaction": {
        "from": "{{account1}}",
        "to": "{{to2}}",
        "value": "0x1FBD0",
        "nonce": 2,
        "gas": 21000
    },
    "labels": {
    	"scenario.id": "{{scenarioID}}"
    }
}
      """
    Then the response code should be 200
    Then I register the following response fields
      | alias        | path |
      | txTwoJobUUID | uuid |
    When I send "POST" request to "{{global.api}}/jobs" with json:
  """
{
    "scheduleUUID": "{{scheduleThreeUUID}}",
	"chainUUID": "{{chain.besu0.UUID}}",
    "type": "eth://ethereum/transaction",
    "transaction": {
        "from": "{{account1}}",
        "to": "{{to3}}",
        "value": "0x1A9C8",
        "gas": 21000
    },
    "labels": {
    	"scenario.id": "{{scenarioID}}"
    }
}
      """
    Then the response code should be 200
    Then I register the following response fields
      | alias          | path |
      | txThreeJobUUID | uuid |
    Then I track the following envelopes
      | ID                    |
      | {{scheduleOneUUID}}   |
      | {{scheduleTwoUUID}}   |
      | {{scheduleThreeUUID}} |
    When I send "PUT" request to "{{global.api}}/jobs/{{txOneJobUUID}}/start"
    Then the response code should be 202
    When I send "PUT" request to "{{global.api}}/jobs/{{txTwoJobUUID}}/start"
    Then the response code should be 202
    When I send "PUT" request to "{{global.api}}/jobs/{{txThreeJobUUID}}/start"
    Then the response code should be 202
    Then Envelopes should be in topic "tx.decoded"
    And Envelopes should have the following fields
      | Receipt.Status | Nonce |
      | 1              | 0     |
      | 1              | 2     |
      | 1              | 1     |

  Scenario: Chaotic nonce
    Given I register the following alias
      | alias | value              |
      | to1   | {{random.account}} |
      | to2   | {{random.account}} |
      | to3   | {{random.account}} |
    Then I set the headers
      | Key         | Value                |
      | X-API-KEY   | {{global.api-key}}   |
      | X-TENANT-ID | {{tenant1.tenantID}} |
    When I send "POST" request to "{{global.api}}/schedules" with json:
      """
{}
      """
    Then the response code should be 200
    Then I register the following response fields
      | alias           | path |
      | scheduleOneUUID | uuid |
    When I send "POST" request to "{{global.api}}/schedules" with json:
      """
{}
      """
    Then the response code should be 200
    Then I register the following response fields
      | alias           | path |
      | scheduleTwoUUID | uuid |
    When I send "POST" request to "{{global.api}}/schedules" with json:
      """
{}
      """
    Then the response code should be 200
    Then I register the following response fields
      | alias             | path |
      | scheduleThreeUUID | uuid |
    When I send "POST" request to "{{global.api}}/jobs" with json:
  """
{
    "scheduleUUID": "{{scheduleOneUUID}}",
	"chainUUID": "{{chain.besu0.UUID}}",
    "type": "eth://ethereum/transaction",
    "transaction": {
        "from": "{{account3}}",
        "to": "{{to1}}",
        "value": "0x186A0",
        "gasPrice": "0x3B9ACA00",
        "gas": 21000
    },
    "labels": {
    	"scenario.id": "{{scenarioID}}"
    }
}
      """
    Then the response code should be 200
    Then I register the following response fields
      | alias        | path |
      | txOneJobUUID | uuid |
    When I send "POST" request to "{{global.api}}/jobs" with json:
  """
{
    "scheduleUUID": "{{scheduleTwoUUID}}",
	"chainUUID": "{{chain.besu0.UUID}}",
    "type": "eth://ethereum/transaction",
    "transaction": {
        "from": "{{account3}}",
        "to": "{{to2}}",
        "value": "0x186A0",
        "nonce": 1,
        "gasPrice": "0x3B9ACA00",
        "gas": 21000
    },
    "labels": {
    	"scenario.id": "{{scenarioID}}"
    }
}
      """
    Then the response code should be 200
    Then I register the following response fields
      | alias        | path |
      | txTwoJobUUID | uuid |
    When I send "POST" request to "{{global.api}}/jobs" with json:
  """
{
    "scheduleUUID": "{{scheduleThreeUUID}}",
	"chainUUID": "{{chain.besu0.UUID}}",
    "type": "eth://ethereum/transaction",
    "transaction": {
        "from": "{{account3}}",
        "to": "{{to3}}",
        "value": "0x186A0",
        "nonce": 0,
        "gasPrice": "0x3B9ACA00",
        "gas": 21000
    },
    "labels": {
    	"scenario.id": "{{scenarioID}}"
    }
}
      """
    Then the response code should be 200
    Then I register the following response fields
      | alias          | path |
      | txThreeJobUUID | uuid |
    Then I track the following envelopes
      | ID                    |
      | {{scheduleOneUUID}}   |
      | {{scheduleTwoUUID}}   |
      | {{scheduleThreeUUID}} |
    When I send "PUT" request to "{{global.api}}/jobs/{{txOneJobUUID}}/start"
    Then the response code should be 202
    When I send "PUT" request to "{{global.api}}/jobs/{{txTwoJobUUID}}/start"
    Then the response code should be 202
    When I send "PUT" request to "{{global.api}}/jobs/{{txThreeJobUUID}}/start"
    Then the response code should be 202
    Then Envelopes should be in topic "tx.decoded"
    And Envelopes should have the following fields
      | Receipt.Status | Nonce | To      |
      | 1              | 0     | {{to1}} |
      | 1              | 1     | {{to2}} |
      | 1              | 2     | {{to3}} |
    When I send "GET" request to "{{global.api}}/jobs/{{txThreeJobUUID}}"
    Then the response code should be 200
    And Response should have the following fields
      | status | logs[0].status | logs[1].status | logs[2].status | logs[3].status | logs[4].status | logs[5].status |
      | MINED  | CREATED        | STARTED        | PENDING        | RECOVERING     | PENDING        | MINED          |

  @private-tx
  Scenario: Private Transaction, invalid nonce order
    Given I register the following alias
      | alias | value              |
      | to1   | {{random.account}} |
    Then I set the headers
      | Key         | Value                |
      | X-API-KEY   | {{global.api-key}}   |
      | X-TENANT-ID | {{tenant1.tenantID}} |
    When I send "POST" request to "{{global.api}}/schedules" with json:
      """
{}
      """
    Then the response code should be 200
    Then I register the following response fields
      | alias           | path |
      | scheduleOneUUID | uuid |
    When I send "POST" request to "{{global.api}}/schedules" with json:
      """
{}
      """
    Then the response code should be 200
    Then I register the following response fields
      | alias           | path |
      | scheduleTwoUUID | uuid |
    When I send "POST" request to "{{global.api}}/schedules" with json:
      """
{}
      """
    Then the response code should be 200
    Then I register the following response fields
      | alias             | path |
      | scheduleThreeUUID | uuid |
    When I send "POST" request to "{{global.api}}/jobs" with json:
  """
{
    "scheduleUUID": "{{scheduleOneUUID}}",
	"chainUUID": "{{chain.besu0.UUID}}",
    "type": "eth://eea/privateTransaction",
    "transaction": {
        "from": "{{account1}}",
        "to": "{{to1}}",
        "nonce": 0,
        "privateFrom": "{{global.nodes.besu[0].privateAddress[0]}}",
        "privateFor": ["{{global.nodes.besu[1].privateAddress[0]}}"]
    },
    "labels": {
    	"scenario.id": "{{scenarioID}}"
    }
}
      """
    Then the response code should be 200
    Then I register the following response fields
      | alias        | path |
      | txOneJobUUID | uuid |
    When I send "POST" request to "{{global.api}}/jobs" with json:
  """
{
    "scheduleUUID": "{{scheduleTwoUUID}}",
	"chainUUID": "{{chain.besu0.UUID}}",
    "type": "eth://eea/privateTransaction",
    "transaction": {
        "from": "{{account1}}",
        "to": "{{to1}}",
        "privateFrom": "{{global.nodes.besu[0].privateAddress[0]}}",
        "privateFor": ["{{global.nodes.besu[1].privateAddress[0]}}"]
    },
    "labels": {
    	"scenario.id": "{{scenarioID}}"
    }
}
      """
    Then the response code should be 200
    Then I register the following response fields
      | alias        | path |
      | txTwoJobUUID | uuid |
    When I send "POST" request to "{{global.api}}/jobs" with json:
  """
{
    "scheduleUUID": "{{scheduleThreeUUID}}",
	"chainUUID": "{{chain.besu0.UUID}}",
    "type": "eth://eea/privateTransaction",
    "transaction": {
        "from": "{{account1}}",
        "to": "{{to1}}",
        "privateFrom": "{{global.nodes.besu[0].privateAddress[0]}}",
        "privateFor": ["{{global.nodes.besu[1].privateAddress[0]}}"]
    },
    "labels": {
    	"scenario.id": "{{scenarioID}}"
    }
}
      """
    Then the response code should be 200
    Then I register the following response fields
      | alias          | path |
      | txThreeJobUUID | uuid |
    Then I track the following envelopes
      | ID                    |
      | {{scheduleOneUUID}}   |
      | {{scheduleTwoUUID}}   |
      | {{scheduleThreeUUID}} |
    When I send "PUT" request to "{{global.api}}/jobs/{{txOneJobUUID}}/start"
    Then the response code should be 202
    When I send "PUT" request to "{{global.api}}/jobs/{{txTwoJobUUID}}/start"
    Then the response code should be 202
    When I send "PUT" request to "{{global.api}}/jobs/{{txThreeJobUUID}}/start"
    Then the response code should be 202
    Given I sleep "5s"
    When I send "GET" request to "{{global.api}}/jobs/{{txOneJobUUID}}"
    Then the response code should be 200
    And Response should have the following fields
      | status | logs[0].status | logs[1].status | logs[2].status | transaction.nonce |
      | STORED | CREATED        | STARTED        | STORED         | 0                 |
    When I send "GET" request to "{{global.api}}/jobs/{{txTwoJobUUID}}"
    Then the response code should be 200
    And Response should have the following fields
      | status | logs[0].status | logs[1].status | logs[2].status | transaction.nonce |
      | STORED | CREATED        | STARTED        | STORED         | 1                 |
    When I send "GET" request to "{{global.api}}/jobs/{{txThreeJobUUID}}"
    Then the response code should be 200
    And Response should have the following fields
      | status | logs[0].status | logs[1].status | logs[2].status | transaction.nonce |
      | STORED | CREATED        | STARTED        | STORED         | 2                 |

#  @geth
#  Scenario: Speed up transaction
#    Given I register the following alias
#      | alias            | value           |
#      | txOneID          | {{random.uuid}} |
#      | gethContractTxID | {{random.uuid}} |
#    Given I set the headers
#      | Key         | Value                |
#      | X-API-KEY   | {{global.api-key}}   |
#      | X-TENANT-ID | {{tenant1.tenantID}} |
#    When I send "POST" request to "{{global.api}}/transactions/deploy-contract" with json:
#      """
#      {
#        "chain": "{{chain.geth0.Name}}",
#        "params": {
#          "contractName": "SimpleToken",
#          "from": "{{account1}}",
#          "maxFeePerGas": "0x5950101"
#        },
#        "labels": {
#          "scenario.id": "{{scenarioID}}",
#          "id": "{{gethContractTxID}}"
#        }
#      }
#      """
#    Then the response code should be 202
#    Then I register the following response fields
#      | alias  | path |
#      | txUUID | uuid |
#    Then I sleep "2s"
#    When I send "PUT" request to "{{global.api}}/transactions/{{txUUID}}/speed-up?boost=0.2"
#    Then the response code should be 200
#    And Response should have the following fields
#      | jobs[1].transaction.data | jobs[1].transaction.maxPriorityFeePerGas |
#      | ~                        | 0x6B2CE01                                |
#    Then I register the following response fields
#      | alias        | path           |
#      | jobs[1].uuid | speedUpJobUUID |
#    Then Envelopes should be in topic "tx.sender"
#    When I send "GET" request to "{{global.api}}/jobs/{{speedUpJobUUID}}"
#    Then the response code should be 200
#    And Response should have the following fields
#      | status | logs[0].status | logs[1].status | logs[2].status | logs[3].status | transaction.hash |
#      | MINED  | CREATED        | STARTED        | PENDING        | MINED          | ~                |
#
#  @geth
#  Scenario: Call off transaction
#    Given I register the following alias
#      | alias            | value           |
#      | txOneID          | {{random.uuid}} |
#      | gethContractTxID | {{random.uuid}} |
#    Given I set the headers
#      | Key         | Value                |
#      | X-API-KEY   | {{global.api-key}}   |
#      | X-TENANT-ID | {{tenant1.tenantID}} |
#    When I send "POST" request to "{{global.api}}/transactions/deploy-contract" with json:
#      """
#      {
#        "chain": "{{chain.geth0.Name}}",
#        "params": {
#          "contractName": "SimpleToken",
#          "from": "{{account1}}",
#          "maxFeePerGas": "0x5B4C273"
#        },
#        "labels": {
#          "scenario.id": "{{scenarioID}}",
#          "id": "{{gethContractTxID}}"
#        }
#      }
#      """
#    Then the response code should be 202
#    Then I register the following response fields
#      | alias  | path |
#      | txUUID | uuid |
#    Then I sleep "2s"
#    When I send "PUT" request to "{{global.api}}/transactions/{{txUUID}}/call-off"
#    Then the response code should be 200
#    And Response should have the following fields
#      | jobs[1].transaction.data | jobs[1].transaction.maxPriorityFeePerGas |
#      | -                        | 0x646D5E4                                |
#    Then I register the following response fields
#      | alias        | path           |
#      | jobs[1].uuid | callOffJobUUID |
#    Then Envelopes should be in topic "tx.sender"
#    When I send "GET" request to "{{global.api}}/jobs/{{callOffJobUUID}}"
#    Then the response code should be 200
#    And Response should have the following fields
#      | status | logs[0].status | logs[1].status | logs[2].status | logs[3].status | transaction.hash |
#      | MINED  | CREATED        | STARTED        | PENDING        | MINED          | ~                |
