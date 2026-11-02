# Builds one layer with gems for each Ruby version on both x86_64 and arm64.

FROM --platform=linux/amd64 public.ecr.aws/sam/build-ruby3.3:1.171.0@sha256:4bf1cea62e6d7facdfd9228ae19e1d575b3fb5ce395826e3dafeef03838b530c AS ruby3.3-amd64
COPY layer/Gemfile layer/Gemfile.lock install-gems.sh /build/
RUN /build/install-gems.sh

FROM --platform=linux/arm64 public.ecr.aws/sam/build-ruby3.3:1.171.0@sha256:4bf1cea62e6d7facdfd9228ae19e1d575b3fb5ce395826e3dafeef03838b530c AS ruby3.3-arm64
COPY layer/Gemfile layer/Gemfile.lock install-gems.sh /build/
RUN /build/install-gems.sh

FROM --platform=linux/amd64 public.ecr.aws/sam/build-ruby3.4:1.171.0@sha256:42eb2365cbbd3fcbf997be3e30ed98ca82223f2b0fd7dc909c1fbc51c9e7785b AS ruby3.4-amd64
COPY layer/Gemfile layer/Gemfile.lock install-gems.sh /build/
RUN /build/install-gems.sh

FROM --platform=linux/arm64 public.ecr.aws/sam/build-ruby3.4:1.171.0@sha256:42eb2365cbbd3fcbf997be3e30ed98ca82223f2b0fd7dc909c1fbc51c9e7785b AS ruby3.4-arm64
COPY layer/Gemfile layer/Gemfile.lock install-gems.sh /build/
RUN /build/install-gems.sh

FROM --platform=linux/amd64 public.ecr.aws/sam/build-ruby4.0:1.171.0@sha256:ccb7be2926d51287111795068a82060d8fe8dc9da49ea78e1813c7ceb8d8d9fa AS ruby4.0-amd64
COPY layer/Gemfile layer/Gemfile.lock install-gems.sh /build/
RUN /build/install-gems.sh

FROM --platform=linux/arm64 public.ecr.aws/sam/build-ruby4.0:1.171.0@sha256:ccb7be2926d51287111795068a82060d8fe8dc9da49ea78e1813c7ceb8d8d9fa AS ruby4.0-arm64
COPY layer/Gemfile layer/Gemfile.lock install-gems.sh /build/
RUN /build/install-gems.sh

FROM public.ecr.aws/sam/build-ruby3.3:1.171.0@sha256:4bf1cea62e6d7facdfd9228ae19e1d575b3fb5ce395826e3dafeef03838b530c
WORKDIR /build
COPY --from=ruby3.3-amd64 /gems/ruby/ amd64/
COPY --from=ruby3.4-amd64 /gems/ruby/ amd64/
COPY --from=ruby4.0-amd64 /gems/ruby/ amd64/
COPY --from=ruby3.3-arm64 /gems/ruby/ arm64/
COPY --from=ruby3.4-arm64 /gems/ruby/ arm64/
COPY --from=ruby4.0-arm64 /gems/ruby/ arm64/

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
