#!/bin/bash

keyword="$1"

# Get the process IDs of all processes matching the keyword
pids=$(pgrep -f "$keyword")

if [ -z "$pids" ]; then
	echo "No processes found matching the keyword '$keyword'"
else
	# Kill each process
	for pid in $pids; do
		echo "Killing process: $pid"
		kill "$pid"
	done
	echo "All processes matching the keyword '$keyword' have been killed."
fi

