#!/usr/bin/env bash

output_directory="$1"

for ((i=2; i<"$#"; i+=2)); do
	j="$((i+1))"

	destination="$output_directory/${!i}"

	mkdir -p "$destination"
	cp "${!j}" "$destination"
done
