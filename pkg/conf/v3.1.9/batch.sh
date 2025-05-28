#!/bin/bash

# Get all manifest files.
# Using find is generally safer than ls for scripting, especially with spaces or special characters.
# However, given your specific filenames, ls -1 manifest-*.yaml is also fine.
find . -maxdepth 1 -name 'manifest-*.yaml' -print0 | while IFS= read -r -d $'\0' filepath; do
  # Get just the filename from the path (e.g., remove "./")
  filename="${filepath##*/}"
  output_file=""

  if [[ "$filename" == "manifest-simple-all-v319.yaml" ]]; then
    # Handle manifest-simple-all-v319.yaml
    # Remove "manifest-" prefix and ".yaml" suffix
    core_part="${filename#manifest-}"        # simple-all-v319.yaml
    core_part="${core_part%.yaml}"           # simple-all-v319
    output_file="taichu-universal-${core_part}.tar.gz" # taichu-universal-simple-all-v319.tar.gz

  elif [[ "$filename" == "manifest-all-v319.yaml" ]]; then
    # Handle manifest-all-v319.yaml
    # Remove "manifest-" prefix and ".yaml" suffix
    core_part="${filename#manifest-}"        # all-v319.yaml
    core_part="${core_part%.yaml}"           # all-v319
    output_file="taichu-universal-${core_part}.tar.gz" # taichu-universal-all-v319.tar.gz

  elif [[ "$filename" == manifest-v*.yaml ]]; then # For files like manifest-v1.29.15.yaml
    # Original logic for versioned files
    version_part="${filename#manifest-}" # v1.29.15.yaml
    version="${version_part%.yaml}"      # v1.29.15
    output_file="taichu-universal-${version}-v319.tar.gz" # taichu-universal-v1.29.15-v319.tar.gz
  else
    echo "WARNING: Skipping unrecognized file format: $filename"
    continue # Skip to the next file
  fi

  if [[ -n "$output_file" ]]; then
    echo "Input: ${filepath}, Output: ${output_file}"
    # Uncomment the line below to run for real
    # ./kk artifact export -m "${filepath}" -o "${output_file}"
  fi
done

echo "Processing complete."