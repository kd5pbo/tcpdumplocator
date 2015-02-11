#!/bin/ksh
# build.sh
# Builds Linux, OS X, and Windows binaries
# by J. Stuart McMurray
# created 20150211
# last modified 20150211
# 
# Copyright (c) 2015 J. Stuart McMurray <kd5pbo@gmail.com>
# 
# Permission to use, copy, modify, and distribute this software for any
# purpose with or without fee is hereby granted, provided that the above
# copyright notice and this permission notice appear in all copies.
# 
# THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
# WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
# MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
# ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
# WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
# ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
# OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.

set -e

GOARCH=amd64

for GOOS in windows linux darwin; do
        O=tcpdumplocator.$GOOS.amd64
        if [[ "$GOOS" == "windows" ]]; then
                O=$O.exe
        fi
        echo Building $O
        go build -o $O  tcpdumplocator.go
done

echo Done.
