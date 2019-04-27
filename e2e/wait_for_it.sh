#!/usr/bin/env sh

set -eo

wait_for_it() {
	local host=$1
	local port=$2
	local result

	while true 
	do
		nc -z $host $port
		result=$?
		if [ $result -eq 0 ]
		then
			exit 0
		fi

		sleep 1
	done
}

wait_for_it $@ 
