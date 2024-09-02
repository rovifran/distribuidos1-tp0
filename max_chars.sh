#!/bin/bash

if [ $# -eq 0 ]; then
    echo "Usage: $0 <csv_file>"
    exit 1
fi

max_length=0
longest_line_number=0

while IFS= read -r line; do
    ((line_number++))
    current_length=$(echo "$line" | wc -c)
    
    if [ $current_length -gt $max_length ]; then
        max_length=$current_length
        longest_line_number=$line_number
    fi
done < "$1"

echo "Longest line is on line $longest_line_number with ${max_length} characters."