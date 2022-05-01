#!/bin/bash

check () {
	RES=`cat $1 | grep "$2" | wc -l`
	if [ $RES -ne $3 ]; then
		echo "$2 expected=$3 actual $RES"
		exit 1
	else
		echo "$2 OK"
	fi
}