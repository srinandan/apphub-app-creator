#!/bin/sh
# Copyright 2025 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.


set -e

# Determines the operating system.
OS="$(uname)"
if [ "${OS}" = "Darwin" ] ; then
  OSEXT="Darwin"
else
  OSEXT="Linux"
fi

# Determine the latest apphub-app-creator version by version number ignoring alpha, beta, and rc versions.
if [ "${APPHUB_APP_CREATOR_VERSION}" = "" ] ; then
  APPHUB_APP_CREATOR_VERSION="$(curl -si  https://api.github.com/repos/srinandan/apphub-app-creator/releases/latest | grep tag_name | sed -E 's/.*"([^"]+)".*/\1/')"
fi

LOCAL_ARCH=$(uname -m)
if [ "${TARGET_ARCH}" ]; then
    LOCAL_ARCH=${TARGET_ARCH}
fi

case "${LOCAL_ARCH}" in
  x86_64|amd64)
    APPHUB_APP_CREATOR_ARCH=x86_64
    ;;
  arm64|armv8*|aarch64*)
    APPHUB_APP_CREATOR_ARCH=arm64
    ;;
  *)
    echo "This system's architecture, ${LOCAL_ARCH}, isn't supported"
    exit 1
    ;;
esac

if [ "${APPHUB_APP_CREATOR_VERSION}" = "" ] ; then
  printf "Unable to get latest apphub-app-creator version. Set APPHUB_APP_CREATOR_VERSION env var and re-run. For example: export APPHUB_APP_CREATOR_VERSION=v1.1"
  exit 1;
fi

APPHUB_APP_CREATOR_FILE=~/.apphub-app-creator
if [ -f "$APPHUB_APP_CREATOR_FILE" ]; then
    rm ${APPHUB_APP_CREATOR_FILE}
fi

# Downloads the apphub-app-creator binary archive.
tmp=$(mktemp -d /tmp/apphub-app-creator.XXXXXX)
NAME="apphub-app-creator_$APPHUB_APP_CREATOR_VERSION"

cd "$tmp" || exit
URL="https://github.com/srinandan/apphub-app-creator/releases/download/${APPHUB_APP_CREATOR_VERSION}/apphub-app-creator_${APPHUB_APP_CREATOR_VERSION}_${OSEXT}_${APPHUB_APP_CREATOR_ARCH}.zip"
SIG_URL="https://github.com/srinandan/apphub-app-creator/releases/download/${APPHUB_APP_CREATOR_VERSION}/apphub-app-creator_${APPHUB_APP_CREATOR_VERSION}_${OSEXT}_${APPHUB_APP_CREATOR_ARCH}.zip.sig"
COSIGN_PUBLIC_KEY="https://raw.githubusercontent.com/srinandan/apphub-app-creator/main/cosign.pub"

download_cli() {
  printf "\nDownloading %s from %s ...\n" "$NAME" "$URL"
  if ! curl -o /dev/null -sIf "$URL"; then
    printf "\n%s is not found, please specify a valid APPHUB_APP_CREATOR_VERSION and TARGET_ARCH\n" "$URL"
    exit 1
  fi
  curl -fsLO "$URL"
  filename="apphub-app-creator_${APPHUB_APP_CREATOR_VERSION}_${OSEXT}_${APPHUB_APP_CREATOR_ARCH}.zip"
  # Check if cosign is installed
  set +e # disable exit on error
  # cosign version 2>&1 >/dev/null
  # RESULT=$?
  # set -e # re-enable exit on error
  # if [ $RESULT -eq 0 ]; then
  #   echo "Verifying the signature of the binary " "$filename"
  #   echo "Downloading the cosign public key"
  #   curl -fsLO -H 'Cache-Control: no-cache, no-store' "$COSIGN_PUBLIC_KEY"
  #   echo "Downloading the signature file " "$SIG_URL"
  #   curl -fsLO -H 'Cache-Control: no-cache, no-store' "$SIG_URL"
  #   sig_filename="apphub-app-creator_${APPHUB_APP_CREATOR_VERSION}_${OSEXT}_${APPHUB_APP_CREATOR_ARCH}.zip.sig"
  #   echo "Verifying the signature"
  #   cosign verify-blob --key "$tmp/cosign.pub" --signature "$tmp/$sig_filename" "$tmp/$filename"
  #   rm "$tmp/$sig_filename"
  #   rm $tmp/cosign.pub
  # else
  #   echo "cosign is not installed, skipping signature verification"
  # fi
  unzip "${filename}"
  rm "${filename}"
}


download_cli

printf ""
printf "\napphub-app-creator %s Download Complete!\n" "$APPHUB_APP_CREATOR_VERSION"
printf "\n"
printf "apphub-app-creator has been successfully downloaded into the %s folder on your system.\n" "$tmp"
printf "\n"

# setup apphub-app-creator
cd "$HOME" || exit
mkdir -p "$HOME/.apphub-app-creator/bin"
mv "${tmp}/apphub-app-creator_${APPHUB_APP_CREATOR_VERSION}_${OSEXT}_${APPHUB_APP_CREATOR_ARCH}/apphub-app-creator" "$HOME/.apphub-app-creator/bin"
mv "${tmp}/apphub-app-creator_${APPHUB_APP_CREATOR_VERSION}_${OSEXT}_${APPHUB_APP_CREATOR_ARCH}/LICENSE.txt" "$HOME/.apphub-app-creator/LICENSE.txt"

printf "Copied apphub-app-creator into the $HOME/.apphub-app-creator/bin folder.\n"
chmod +x "$HOME/.apphub-app-creator/bin/apphub-app-creator"
rm -r "${tmp}"

# Print message
printf "\n"
printf "Added the apphub-app-creator to your path with:"
printf "\n"
printf "  export PATH=\$PATH:\$HOME/.apphub-app-creator/bin \n"
printf "\n"

export PATH=$PATH:$HOME/.apphub-app-creator/bin
