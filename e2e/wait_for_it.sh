#!/usr/bin/env sh


usage() {
	if [ $# -ne 2 ]
	then
		echo "$#"
		echo "usage: wait_for_it host port "
		exit -1
	fi
}

wait_for_it() {
	usage $@

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
