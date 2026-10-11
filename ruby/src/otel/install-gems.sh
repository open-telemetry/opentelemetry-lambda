#!/bin/sh
# Installs the layer's gems for the Ruby version and architecture of the
# current build image into /gems/ruby/<X.Y.0>, keeping only what ships.
set -eu

# Without this RubyGems also copies compiled extensions (e.g. bigdecimal.so)
# into gems/<name>/lib/, an arch-agnostic path that would clash between
# amd64 and arm64. The copy under extensions/<arch>-linux/ is enough.
echo "install_extension_in_lib: false" > ~/.gemrc

# BUNDLE_VERSION=system: use the image's Bundler instead of installing the
# lockfile's "BUNDLED WITH" version into the layer.
cd /build
BUNDLE_PATH=/gems BUNDLE_FROZEN=true BUNDLE_VERSION=system \
  bundle install --jobs "$(nproc)"

# Drop download cache, docs, build leftovers, binstubs and build logs.
cd /gems/ruby/*/
rm -rf cache doc build_info bin plugins gems/*/ext
find extensions \( -name mkmf.log -o -name gem_make.out \) -delete
