# Builds one layer with gems for Ruby 3.3 and 3.4 on both x86_64 and arm64.
# RubyGems only activates gem specs whose platform matches the running
# machine, so both architectures' gems can live side by side in
# ruby/gems/<X.Y.0>. The four install stages below are independent, so
# BuildKit runs them in parallel (arm64 via QEMU on amd64 hosts and vice versa).

FROM --platform=linux/amd64 public.ecr.aws/sam/build-ruby3.3:latest@sha256:0e8b401558ba3d59722ea169d0da854631136a23eded32109d6a8ed2345ab0ee AS ruby3.3-amd64
COPY layer/Gemfile layer/Gemfile.lock install-gems.sh /build/
RUN /build/install-gems.sh

FROM --platform=linux/arm64 public.ecr.aws/sam/build-ruby3.3:latest@sha256:0e8b401558ba3d59722ea169d0da854631136a23eded32109d6a8ed2345ab0ee AS ruby3.3-arm64
COPY layer/Gemfile layer/Gemfile.lock install-gems.sh /build/
RUN /build/install-gems.sh

FROM --platform=linux/amd64 public.ecr.aws/sam/build-ruby3.4:latest@sha256:a208c01413754ea473357ffdc850fab3cdadf6dac776a0566242464dba2aa02b AS ruby3.4-amd64
COPY layer/Gemfile layer/Gemfile.lock install-gems.sh /build/
RUN /build/install-gems.sh

FROM --platform=linux/arm64 public.ecr.aws/sam/build-ruby3.4:latest@sha256:a208c01413754ea473357ffdc850fab3cdadf6dac776a0566242464dba2aa02b AS ruby3.4-arm64
COPY layer/Gemfile layer/Gemfile.lock install-gems.sh /build/
RUN /build/install-gems.sh

# Final stage only merges and zips, so it runs on the host's native platform.
FROM public.ecr.aws/sam/build-ruby3.3:latest@sha256:0e8b401558ba3d59722ea169d0da854631136a23eded32109d6a8ed2345ab0ee
WORKDIR /build
COPY --from=ruby3.3-amd64 /gems/ruby/ amd64/
COPY --from=ruby3.4-amd64 /gems/ruby/ amd64/
COPY --from=ruby3.3-arm64 /gems/ruby/ arm64/
COPY --from=ruby3.4-arm64 /gems/ruby/ arm64/

# Paths that exist for both architectures must be identical, otherwise one
# would overwrite the other in the merged layer. Fail if a (future) native
# gem brings such a clash back.
RUN if diff -rq amd64 arm64 2>&1 | grep -v '^Only in '; then \
      echo "ERROR: amd64 and arm64 gem trees clash on the paths above" >&2; exit 1; \
    fi

RUN mkdir -p ruby/gems \
    && cp -a amd64/. ruby/gems/ \
    && cp -a arm64/. ruby/gems/

COPY layer/otel-handler ./
COPY layer/wrapper.rb ./
RUN zip -qr opentelemetry-ruby-layer.zip ruby/ otel-handler wrapper.rb

CMD ["cp", "/build/opentelemetry-ruby-layer.zip", "/out/opentelemetry-ruby-layer.zip"]
