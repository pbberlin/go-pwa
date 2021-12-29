#!/bin/sh

# <https://developers.google.com/speed/webp/docs/precompiled>

export PATH="/c/Program Files/libwebp/bin/"

cd ../app-bucket/img


echo "start"

for file in *.png; do 
    if [ -f "$file" ]; then 
        echo "file is $file";
        base="${file%.*}"; 
        echo $base;
        # cwebp "$file" -o "$file.webp";
        cwebp $file -o "${base}.webp";
    fi 
done

echo "stop"

