#!/bin/bash


# BSD 3-Clause License

# Copyright (c) 2023, Daniel Gehrer, xa1.at
# All rights reserved.
# 
# Redistribution and use in source and binary forms, with or without
# modification, are permitted provided that the following conditions are met:
# 
# 1. Redistributions of source code must retain the above copyright notice, this
#    list of conditions and the following disclaimer.
# 
# 2. Redistributions in binary form must reproduce the above copyright notice,
#    this list of conditions and the following disclaimer in the documentation
#    and/or other materials provided with the distribution.
# 
# 3. Neither the name of the copyright holder nor the names of its
#    contributors may be used to endorse or promote products derived from
#    this software without specific prior written permission.
# 
# THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
# AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
# IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
# DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE
# FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
# DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
# SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
# CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
# OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
# OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

# https://github.com/xa17d/serve/blob/main/build.sh

go get
set -e
rm -rf build

TARGETS=(
    "linux/386/"
    "linux/amd64/"
    "linux/arm64/"
    "darwin/amd64/"
    "darwin/arm64/"
    "windows/386/.exe"
    "windows/amd64/.exe"
)

product_name="payshop3"
start_dir="${PWD}"

for target in "${TARGETS[@]}"; do
    # Split the target operating system and architecture into separate variables
    IFS='/' read -r os arch extension <<< "${target}"

    # Define the output directory
    output_dir="build/${os}_${arch}"
    output_name="${product_name}${extension}"

    # Create the output directory if it doesn't exist
    mkdir -p "${output_dir}"

    # Build the program for the target operating system and architecture
    GOOS="${os}" GOARCH="${arch}" go build -ldflags="-s -w" -trimpath -o "${output_dir}/${output_name}" .

    # Zip the executable file
    cd "${output_dir}"
    zip "build.zip" "${output_name}"
    mv "build.zip" "../${product_name}_${os}_${arch}.zip"

    cd "${start_dir}"

    # Print a message indicating that the build was successful
    echo "Build for ${os}/${arch} successful."
done