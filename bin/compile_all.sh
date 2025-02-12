#!/bin/bash

# Set the source and destination directories
SRC_DIR="bin/source_code"
DEST_DIR="bin"

# Compile each .c file in the source directory
for file in "$SRC_DIR"/*.c; do
    # Get the filename without extension
    filename=$(basename "$file" .c)
    
    # Compile the file and place the binary in the destination directory
    gcc "$file" -o "$DEST_DIR/$filename"
    
    # Check if the compilation was successful
    if [ $? -eq 0 ]; then
        echo "Compiled $file -> $DEST_DIR/$filename"
    else
        echo "Failed to compile $file"
    fi
done
