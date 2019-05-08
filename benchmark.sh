#!/bin/sh

endpoint=https://${proj}.appspot.com

uuid() {
    uuidgen | tr -d "\-\n"
}

trace_on() {
    echo "X-Cloud-Trace-Context:$(uuid)/0;o=1"
}

trace_off() {
    echo "X-Cloud-Trace-Context:$(uuid)/0;o=0"
}

curl ${endpoint}/_/health --header $(trace_off)
curl ${endpoint}/cloud/datastore/put --header $(trace_on)
curl ${endpoint}/cloud/datastore/get --header $(trace_on)
curl ${endpoint}/cloud/datastore/txgetput --header $(trace_on)
curl ${endpoint}/appengine/datastore/put --header $(trace_on)
curl ${endpoint}/appengine/datastore/get --header $(trace_on)
curl ${endpoint}/appengine/datastore/txgetput --header $(trace_on)
curl ${endpoint}/cloud/tasks --header $(trace_on)
curl ${endpoint}/appengine/taskqueue --header $(trace_on)
