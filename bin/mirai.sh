#!/bin/bash
function mirai() {
	# Determine the directory where the script is located
	script_dir=$(pwd)

	# Load parameters from config.json in the current directory
	config=$(cat "$script_dir/config.json")

	# Extract the script name from the command-line argument
	script_name=$1

	# Extract the script command from the config
	script_cmd=$(echo "$config" | jq -r ".scripts[\"$script_name\"]")

	# Check if the script command is defined
	if [ "$script_cmd" != "null" ] && [ -n "$script_cmd" ]; then
		# Execute the specified script command
		eval "$script_cmd"
	else
		echo "No script defined for '$script_name' in the config.json file."
	fi
}
