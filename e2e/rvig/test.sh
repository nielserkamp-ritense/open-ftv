#!/bin/bash

APP="mock/datasources/generic/cmd/main.go"
METADATA="testdata/apps/rvig/dataspace"
PORT=9982
URL1="http://localhost:${PORT}/healthz"
URL2="http://localhost:${PORT}/v1/meta/datasource/brp"
URL3="http://localhost:${PORT}/v1/haalcentraal/api/brp/personen"
LOGDIR="/tmp/rvig"
LOG="${LOGDIR}/mock-ds.txt"

finito () {
  kill -s SIGTERM ${PID}
  sleep 2s
  echo "process exited with code $1" >> ${LOG}
  cat ${LOG}
  exit $1
}

mkdir -p ${LOGDIR}
rm -f ${LOG}

go run ${APP} --port=${PORT} --data=${METADATA} >> ${LOG} 2>&1 &

STATUS="000"
until [[ $STATUS != 000 ]]; do
  STATUS=$(curl --head --location --connect-timeout 5 --write-out %{http_code} --silent --output /dev/null ${URL1})
done

PID=$(ps aux | grep ${METADATA} | grep ${PORT} | awk '{print $2}')
# echo "PID = ${PID}" >> ${LOG}

if [[ $STATUS != 200 ]]; then
  echo "failed to start service; healthz status = ${STATUS}" >> ${LOG}
  finito 1
fi

echo "healthy status" >> ${LOG}

STATUS=$(curl --location --connect-timeout 5 --write-out %{http_code} --silent --output /dev/null ${URL2})
if [[ $STATUS != 200 ]]; then
  echo "failed to get metadata; status = ${STATUS}" >> ${LOG}
  finito 1
fi

echo "metadata accessible" >> ${LOG}

STATUS=$(curl --location --connect-timeout 5 --write-out %{http_code} --silent --output /dev/null -X POST ${URL3})
if [[ $STATUS != 200 ]]; then
  echo "failed to get endpoint; status = ${STATUS}" >> ${LOG}
  finito 1
fi

echo "endpoint accessible" >> ${LOG}

finito 0
