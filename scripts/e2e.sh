#!/usr/bin/env bash

set -euf -o pipefail

METRIC_DATA_QUERY_TEMPLATE_FILE="metric-data-queries-template.json"
METRIC_DATA_QUERY_FILE="metric-data-queries.json"
METRIC_DATA_RESULTS_FILE="metric-data-results.json"

export START_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
export END_DATE=$(date -d "+10 minutes" -u +"%Y-%m-%dT%H:%M:%SZ")
export NAMESPACE="e2e_test_${START_DATE}"

echo "INFO  | Running end to end test with:"
echo "INFO  |     START_DATE: ${START_DATE}"
echo "INFO  |     END_DATE:   ${END_DATE}"
echo "INFO  |     NAMESPACE:  ${NAMESPACE}"

echo "INFO  | Running cwmonitor and test help containers"
docker-compose up -d
timeout 180s docker-compose logs -f || true
docker-compose down

echo "INFO  | Generating metric query json"
jq -c --arg v "${NAMESPACE}" \
    '(.[] | .MetricStat | .Metric | .Namespace) |= sub("Test"; $v)' \
    ${METRIC_DATA_QUERY_TEMPLATE_FILE} | tee ${METRIC_DATA_QUERY_FILE}

echo "INFO  | Fetching the metric data uploaded by cwmonitor"
aws cloudwatch get-metric-data \
    --metric-data-queries file://./${METRIC_DATA_QUERY_FILE} \
    --start-time ${START_DATE} \
    --end-time ${END_DATE} | jq -c '.' | tee ${METRIC_DATA_RESULTS_FILE}

EXPECTED_METRICS=$(jq '.[] | .Id' ${METRIC_DATA_QUERY_FILE} | uniq | wc -l)
RESULT_METRICS=$(jq '.MetricDataResults | .[] | select(.Values | length >= 3) | .Id' \
                    ${METRIC_DATA_RESULTS_FILE} | uniq | wc -l)
if [[ "$EXPECTED_METRICS" -ne "$RESULT_METRICS" ]]; then
  echo "Expected ${EXPECTED_METRICS} metrics with 3 data points but found ${RESULT_METRICS}"
  exit 1
fi

if [[ $(jq '.MetricDataResults | .[] | select(.Id | contains("health_healthy")) | .Values | map(select(. == 1)) | length' \
           ${METRIC_DATA_RESULTS_FILE}) -ne 3 ]]; then
    echo "Expected to find 3 healthy data points for healthy container"
    exit 1
fi

if [[ $(jq '.MetricDataResults | .[] | select(.Id | contains("health_unhealthy")) | .Values | map(select(. == 0)) | length' \
           ${METRIC_DATA_RESULTS_FILE}) -ne 3 ]]; then
    echo "Expected to find 3 unhealthy data points for unhealthy container"
    exit 1
fi
