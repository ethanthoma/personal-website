#!/bin/bash

# Load parameters from config.json
config=$(cat "./config.json")

# Source and destination dirs
src_dir=$(echo $config | jq -r ".root")
if [ "$src_dir" = "null" ]; then 
	src_dir="./src/"
fi
dist_dir="./dist/"

# Clear dist directory
echo "Creating distribution dir"
mkdir -p "$dist_dir"

# Construct rsync command
echo "Constructing rsync command"
rsync_cmd="rsync -qzap -m --include '*/'"

# Exclude code that requires compiling
rsync_cmd+=" --exclude '*.scss'"

# Get include and exclude paths
include=$(echo $config | jq -r '.include[]')
exclude=$(echo $config | jq -r '.exclude[]')

# Disable globbing
set -o noglob

# Add exclude options for each exlusion from include
for element in $exclude; do
	rsync_cmd+=" --exclude='$element'"
done

# Add include options for each element in the list
for element in $include; do
	rsync_cmd+=" --include='$element'"
done

# Add source and destination paths
rsync_cmd+=" --exclude '*' '$src_dir' '$dist_dir'"

# Enable globbing
set +o noglob

# Copy source code
echo "Syncing source to distribution..."
eval $rsync_cmd
echo "Done"

# Find all .scss files in src/ and its subdirectories
echo "Compiling scss..."
find "$src_dir" -type f -name '*.scss' -print0 | while IFS= read -r -d '' file; do
	# Get the relative path of the file (without src_dir)
	relative_path="${file#$src_dir}"

	# Destination directory path
	dest_dir="${dist_dir}${relative_path%/*}"

	# Create the destination directory if it doesn't exist
	mkdir -p "$dest_dir"

	# Compile .scss to .css using your compilation command
	# Replace the following line with your compilation command
	sass --no-source-map "$file" "${dist_dir}${relative_path%.scss}.css"
done
echo "Done"

# Delete files that no longer exist in source directory
echo "Deleting obsolete files..."
rsync -r --delete --exclude "*.scss" --exclude "*/" "$src_dir" "$dist_dir"
echo "Done"

echo "Build complete!"
