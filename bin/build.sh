#!/bin/bash

# File extensions to compile
compile_ext=("*.scss")

# Load parameters from config.json
config=$(cat "./config.json")

# Source and destination dirs
src_dir=$(echo $config | jq -r ".root")
if [ "$src_dir" = "null" ]; then 
	src_dir="./src/"
fi

dist_dir=$(echo $config | jq -r ".dist")
if [ "$dist_dir" = "null" ]; then
	dist_dir="./dist/"
fi

# Create dist directory if it doesn't exist
mkdir -p "$dist_dir"

# Disable globbing
set -o noglob

# Exclude files needed to compile
exclude_cmd=""
for item in $compile_ext; do
	exclude_cmd+=" --exclude '$item'"
done

# Get include and exclude paths
include=$(echo $config | jq -r '.include[]')
exclude=$(echo $config | jq -r '.exclude[]')

# Create exlusion files and dirs part of rsync command
for item in $exclude; do
	exclude_cmd+=" --exclude '$item'"
done
                
# Function to check if a file should be excluded
should_include() {
	local file="$1"
	local exclude="$2"  # Remaining arguments are the exclude patterns

	for pattern in "${exclude}"; do
		if [[ "$file" =~ $pattern ]]; then
			return 1
		fi
	done

	return 0
}

# Create and execute rsync command for each dir in include
echo "Building code from source to distribution..."
for item in $include; do
	rsync_cmd="rsync -qzap -m -del --include '*/'"

	# Exclude code that requires compiling
	rsync_cmd+="$exclude_cmd"

	# Separate globs from include item
	filter=$(echo $item | sed "s,^[^\*]*,,")
	src=$(echo $item | sed "s,/\*.*$,," | sed "s,\**,,")

	# Find .scss files within the source directory
	find "$src_dir/$src" -type f -name '*.scss' -print0 | while IFS= read -r -d '' file; do
		# Get the relative path of the file
		relative_path=$(echo "$file" | sed -e "s,^$src_dir/$src/,,")

		# Compile non-excluded code
		if [ -z "$exclude" ] || should_include "$file" "$exclude"; then
			sass --no-source-map "$file" "${dist_dir}${relative_path%.scss}.css"
		fi
	done

	# Add source and destination paths
	rsync_cmd+=" --include '$filter' --exclude '*' '$src_dir/$src/' '$dist_dir'"

	# Execute rsync to copy code
	eval $rsync_cmd
done
echo "Done"

# Enable globbing
set +o noglob

echo "Build complete!"
