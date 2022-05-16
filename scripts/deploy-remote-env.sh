#!/usr/bin/env bash

# Exit on error
set -Ee

TOKEN_HEADER="Circle-Token: ${CIRCLECI_TOKEN}"

#Pass parameters to the Circle CI pipeline
PARAMETERS=""
[ "$ORCHESTRATE_NAMESPACE" ] && export PARAMETERS=$PARAMETERS,\"orchestrate-namespace\":\"$ORCHESTRATE_NAMESPACE\"
[ "$ORCHESTRATE_TAG" ] && export PARAMETERS=$PARAMETERS,\"orchestrate-tag\":\"$ORCHESTRATE_TAG\"
[ "$PARAMETERS" ] && PARAMETERS=${PARAMETERS:1}

echo "Deploying helm charts version $KUBERNETES_BRANCH with parameters: $PARAMETERS"

#Create CircleCI pipeline
RESPONSE=$(curl -s --request POST --header "${TOKEN_HEADER}" --header "Content-Type: application/json" --data '{"branch":"'${KUBERNETES_BRANCH-master}'","parameters":{'${PARAMETERS}'}}' https://circleci.com/api/v2/project/github/ConsenSys/orchestrate-kubernetes/pipeline)
echo $RESPONSE
ID=$(echo $RESPONSE | jq '.id' -r)
NUMBER=$(echo $RESPONSE | jq '.number' -r)

echo "Circle CI pipeline created: $ID"

#Timeout after 4*150 seconds = 10min
SLEEP=4
RETRY=150

for i in $(seq 1 1 $RETRY); do
  sleep $SLEEP

  # Get pipeline status
  STATUS=$(curl -s --request GET --header "${TOKEN_HEADER}" --header "Content-Type: application/json" https://circleci.com/api/v2/pipeline/${ID}/workflow | jq '.items[0].status' -r)
  echo "$i/$RETRY - $STATUS"

  if [[ $STATUS != 'running' ]]; then
    break
  fi

  if [ $i = $RETRY ]; then
    echo "Timeout"
  fi
done

echo "Final status: ${STATUS}"

PIPELINE_ID=$(curl -s --request GET --header "${TOKEN_HEADER}" --header "Content-Type: application/json" https://circleci.com/api/v2/pipeline/${ID}/workflow | jq '.items[0].id' -r)
echo "See the pipeline https://app.circleci.com/pipelines/github/ConsenSys/orchestrate-kubernetes/${NUMBER}/workflows/${PIPELINE_ID}"

if [ "$STATUS" != "success" ]; then
  exit 1
fi
