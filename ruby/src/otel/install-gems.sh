#!/bin/sh
set -eu

# Keep compiled extensions only under extensions/<arch>-linux/, not in the
# arch-agnostic gems/<name>/lib/, so amd64 and arm64 trees can be merged.
echo "install_extension_in_lib: false" > ~/.gemrc

cd /build
BUNDLE_PATH=/gems BUNDLE_FROZEN=true BUNDLE_VERSION=system \
  bundle install --jobs "$(nproc)"

cd /gems/ruby/*/
rm -rf cache doc build_info bin plugins gems/*/ext
find extensions \( -name mkmf.log -o -name gem_make.out \) -delete
