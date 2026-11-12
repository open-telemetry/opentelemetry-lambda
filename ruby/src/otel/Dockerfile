# Builds one layer with gems for each Ruby version on both x86_64 and arm64.

FROM --platform=linux/amd64 public.ecr.aws/sam/build-ruby3.3:1.172.1@sha256:c68d116fc5595c019a20f2ebe6b81ddee31dc13a18a525c1213354f9e1aa7ca1 AS ruby3.3-amd64
COPY layer/Gemfile layer/Gemfile.lock install-gems.sh /build/
RUN /build/install-gems.sh

FROM --platform=linux/arm64 public.ecr.aws/sam/build-ruby3.3:1.172.1@sha256:c68d116fc5595c019a20f2ebe6b81ddee31dc13a18a525c1213354f9e1aa7ca1 AS ruby3.3-arm64
COPY layer/Gemfile layer/Gemfile.lock install-gems.sh /build/
RUN /build/install-gems.sh

FROM --platform=linux/amd64 public.ecr.aws/sam/build-ruby3.4:1.172.1@sha256:f6ce14033547eedf1ed04f4318b0afab884dd8fdd02de1d2de64c67f24f115da AS ruby3.4-amd64
COPY layer/Gemfile layer/Gemfile.lock install-gems.sh /build/
RUN /build/install-gems.sh

FROM --platform=linux/arm64 public.ecr.aws/sam/build-ruby3.4:1.172.1@sha256:f6ce14033547eedf1ed04f4318b0afab884dd8fdd02de1d2de64c67f24f115da AS ruby3.4-arm64
COPY layer/Gemfile layer/Gemfile.lock install-gems.sh /build/
RUN /build/install-gems.sh

FROM --platform=linux/amd64 public.ecr.aws/sam/build-ruby4.0:1.172.1@sha256:a8d15f2ed5117498f13abe6e9768b00ec9be711f510ce3e373bd8ebf6e063b62 AS ruby4.0-amd64
COPY layer/Gemfile layer/Gemfile.lock install-gems.sh /build/
RUN /build/install-gems.sh

FROM --platform=linux/arm64 public.ecr.aws/sam/build-ruby4.0:1.172.1@sha256:a8d15f2ed5117498f13abe6e9768b00ec9be711f510ce3e373bd8ebf6e063b62 AS ruby4.0-arm64
COPY layer/Gemfile layer/Gemfile.lock install-gems.sh /build/
RUN /build/install-gems.sh

FROM public.ecr.aws/sam/build-ruby3.3:1.172.1@sha256:c68d116fc5595c019a20f2ebe6b81ddee31dc13a18a525c1213354f9e1aa7ca1
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
