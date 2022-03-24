@account-management
Feature: Account management
  As as external developer
  I want to generate new accounts and use them to sign transactions

  Background:
    Given I have the following tenants
      | alias   | tenantID        |
      | tenant1 | {{random.uuid}} |
    Given I register the following contracts
      | name        | artifacts        | API-KEY            | Tenant               |
      | SimpleToken | SimpleToken.json | {{global.api-key}} | {{tenant1.tenantID}} |

  Scenario: Import account and update it and sign with it
    Given I register the following alias
      | alias          | value           |
      | sendTxID       | {{random.uuid}} |
      | generateAccID  | {{random.uuid}} |
      | generateAccID2 | {{random.uuid}} |
    Given I have the following account
      | alias     |
      | importAcc |
    Given I set the headers
      | Key         | Value                |
      | X-API-KEY   | {{global.api-key}}   |
      | X-TENANT-ID | {{tenant1.tenantID}} |
    When I send "POST" request to "{{global.api}}/accounts/import" with json:
  """
{
    "alias": "{{generateAccID}}",
    "privateKey": "{{importAcc.private_key}}",
    "attributes": {
    	"scenario_id": "{{scenarioID}}"
    }
}
      """
    Then the response code should be 200
    And Response should have the following fields
      | alias             | attributes.scenario_id | address               |
      | {{generateAccID}} | {{scenarioID}}         | {{importAcc.address}} |
    Then I register the following response fields
      | alias        | path    |
      | importedAddr | address |
    When I send "PATCH" request to "{{global.api}}/accounts/{{importedAddr}}" with json:
  """
{
    "alias": "{{generateAccID2}}",
    "attributes": {
    	"new_attribute": "{{scenarioID}}"
    }
}
      """
    Then the response code should be 200
    Then I send "GET" request to "{{global.api}}/accounts/{{importedAddr}}"
    Then the response code should be 200
    And Response should have the following fields
      | alias              | attributes.new_attribute |
      | {{generateAccID2}} | {{scenarioID}}           |
    Then I track the following envelopes
      | ID           |
      | {{sendTxID}} |
    When I send "POST" request to "{{global.api}}/transactions/deploy-contract" with json:
  """
{
    "chain": "{{chain.besu0.Name}}",
    "params": {
        "contractName": "SimpleToken",
        "from": "{{importedAddr}}"
    },
    "labels": {
    	"scenario.id": "{{scenarioID}}",
    	"id": "{{sendTxID}}"
    }
}
      """
    Then the response code should be 202
    Then Envelopes should be in topic "tx.decoded"

  Scenario: Generate account and send transaction
    Given I register the following alias
      | alias         | value           |
      | generateAccID | {{random.uuid}} |
      | sendTxID      | {{random.uuid}} |
    Given I set the headers
      | Key         | Value                |
      | X-API-KEY   | {{global.api-key}}   |
      | X-TENANT-ID | {{tenant1.tenantID}} |
    When I send "POST" request to "{{global.api}}/accounts" with json:
  """
{
    "alias": "{{generateAccID}}",
    "storeID": "{{global.qkmStoreID}}",
    "attributes": {
    	"scenario_id": "{{scenarioID}}"
    }
}
      """
    Then the response code should be 200
    And Response should have the following fields
      | storeID               |
      | {{global.qkmStoreID}} |
    Then I register the following response fields
      | alias            | path    |
      | generatedAccAddr | address |
    Then I track the following envelopes
      | ID           |
      | {{sendTxID}} |
    Given I set the headers
      | Key         | Value                |
      | X-API-KEY   | {{global.api-key}}   |
      | X-TENANT-ID | {{tenant1.tenantID}} |
    When I send "POST" request to "{{global.api}}/transactions/deploy-contract" with json:
  """
{
    "chain": "{{chain.besu0.Name}}",
    "params": {
        "contractName": "SimpleToken",
        "from": "{{generatedAccAddr}}"
    },
    "labels": {
    	"scenario.id": "{{scenarioID}}",
    	"id": "{{sendTxID}}"
    }
}
      """
    Then the response code should be 202
    Then Envelopes should be in topic "tx.decoded"

  Scenario: Should fail to create an account on a not existing QKM store
    Given I register the following alias
      | alias         | value           |
      | generateAccID | {{random.uuid}} |
    Given I set the headers
      | Key         | Value                |
      | X-API-KEY   | {{global.api-key}}   |
      | X-TENANT-ID | {{tenant1.tenantID}} |
    When I send "POST" request to "{{global.api}}/accounts" with json:
  """
{
    "alias": "{{generateAccID}}",
    "storeID": "not-existing",
    "attributes": {
    	"scenario_id": "{{scenarioID}}"
    }
}
      """
    Then the response code should be 424

  Scenario: Sending transaction with not existed account
    Given I register the following alias
      | alias    | value              |
      | fromAcc  | {{random.account}} |
      | sendTxID | {{random.uuid}}    |
    Then I track the following envelopes
      | ID           |
      | {{sendTxID}} |
    Given I set the headers
      | Key         | Value                |
      | X-API-KEY   | {{global.api-key}}   |
      | X-TENANT-ID | {{tenant1.tenantID}} |
    When I send "POST" request to "{{global.api}}/transactions/deploy-contract" with json:
  """
{
    "chain": "{{chain.besu0.Name}}",
    "params": {
        "contractName": "SimpleToken",
        "from": "{{fromAcc}}"
    },
    "labels": {
    	"scenario.id": "{{scenarioID}}",
    	"id": "{{sendTxID}}"
    }
}
      """
    Then the response code should be 422

