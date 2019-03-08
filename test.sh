#!/bin/bash

for i in {1..1000}
do
        curl -d '{"key": "k'$i'", "value": '$i'}' -X POST http://localhost:8000/set
        curl -d '{"key": "k'$i'"}' -X GET http://localhost:8000/get
        echo ""
done
