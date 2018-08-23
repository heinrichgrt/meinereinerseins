#!/bin/sh
set -
IFS=$'\n'
DEFAULTVALS="initialDefaults.txt"                                      
BASEURL="http://127.0.0.1:2379/v2/keys/trelloprinter/config/default"
for j in `cat $DEFAULTVALS| tr -d "[:blank:]" | grep -v "^//"  `; do
    KEY=$(echo $j | cut -d '='  -f 1)
    VALUE=$(echo $j | cut -d '='  -f 2)
    echo "key $KEY and value $VALUE"
    curl ${BASEURL}/${KEY} -XPUT -d value=${VALUE}

done