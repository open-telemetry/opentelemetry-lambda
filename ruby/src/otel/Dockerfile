# Builds one layer with gems for each Ruby version on both x86_64 and arm64.

FROM --platform=linux/amd64 public.ecr.aws/sam/build-ruby3.3:1.170.0@sha256:2a435b9d7fb00ce6f7faae69025d5e61efab0637830e3a98a7dad5389cd7840d AS ruby3.3-amd64
COPY layer/Gemfile layer/Gemfile.lock install-gems.sh /build/
RUN /build/install-gems.sh

FROM --platform=linux/arm64 public.ecr.aws/sam/build-ruby3.3:1.170.0@sha256:2a435b9d7fb00ce6f7faae69025d5e61efab0637830e3a98a7dad5389cd7840d AS ruby3.3-arm64
COPY layer/Gemfile layer/Gemfile.lock install-gems.sh /build/
RUN /build/install-gems.sh

FROM --platform=linux/amd64 public.ecr.aws/sam/build-ruby3.4:1.170.0@sha256:71e021ba62c04fb8f17f26b895322023b1946426c672407159dd2d21b5ab79e1 AS ruby3.4-amd64
COPY layer/Gemfile layer/Gemfile.lock install-gems.sh /build/
RUN /build/install-gems.sh

FROM --platform=linux/arm64 public.ecr.aws/sam/build-ruby3.4:1.170.0@sha256:71e021ba62c04fb8f17f26b895322023b1946426c672407159dd2d21b5ab79e1 AS ruby3.4-arm64
COPY layer/Gemfile layer/Gemfile.lock install-gems.sh /build/
RUN /build/install-gems.sh

FROM --platform=linux/amd64 public.ecr.aws/sam/build-ruby4.0:1.170.0@sha256:aa59b01b9b2f34e9b3bce43cfb8ff657b25becaf97d3faa0f711d7a24aefe718 AS ruby4.0-amd64
COPY layer/Gemfile layer/Gemfile.lock install-gems.sh /build/
RUN /build/install-gems.sh

FROM --platform=linux/arm64 public.ecr.aws/sam/build-ruby4.0:1.170.0@sha256:aa59b01b9b2f34e9b3bce43cfb8ff657b25becaf97d3faa0f711d7a24aefe718 AS ruby4.0-arm64
COPY layer/Gemfile layer/Gemfile.lock install-gems.sh /build/
RUN /build/install-gems.sh

FROM public.ecr.aws/sam/build-ruby3.3:1.170.0@sha256:2a435b9d7fb00ce6f7faae69025d5e61efab0637830e3a98a7dad5389cd7840d
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
