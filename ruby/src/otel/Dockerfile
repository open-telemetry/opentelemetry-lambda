# Builds one layer with gems for each Ruby version on both x86_64 and arm64.

FROM --platform=linux/amd64 public.ecr.aws/sam/build-ruby3.3:1.174.1@sha256:05414c9464710198a1f15fa14bb4620a54003578aa06b6605c1f497c1f048511 AS ruby3.3-amd64
COPY layer/Gemfile layer/Gemfile.lock install-gems.sh /build/
RUN /build/install-gems.sh

FROM --platform=linux/arm64 public.ecr.aws/sam/build-ruby3.3:1.174.1@sha256:05414c9464710198a1f15fa14bb4620a54003578aa06b6605c1f497c1f048511 AS ruby3.3-arm64
COPY layer/Gemfile layer/Gemfile.lock install-gems.sh /build/
RUN /build/install-gems.sh

FROM --platform=linux/amd64 public.ecr.aws/sam/build-ruby3.4:1.174.1@sha256:d213c92d1bdcb9c0f565c2dcea205c4beff431774de16dee7e86c1a9a9ad803d AS ruby3.4-amd64
COPY layer/Gemfile layer/Gemfile.lock install-gems.sh /build/
RUN /build/install-gems.sh

FROM --platform=linux/arm64 public.ecr.aws/sam/build-ruby3.4:1.174.1@sha256:d213c92d1bdcb9c0f565c2dcea205c4beff431774de16dee7e86c1a9a9ad803d AS ruby3.4-arm64
COPY layer/Gemfile layer/Gemfile.lock install-gems.sh /build/
RUN /build/install-gems.sh

FROM --platform=linux/amd64 public.ecr.aws/sam/build-ruby4.0:1.174.1@sha256:21098addf33f0e9ef7e9e73f64afc1ca918be10e882486d94b9e8c2427426c9e AS ruby4.0-amd64
COPY layer/Gemfile layer/Gemfile.lock install-gems.sh /build/
RUN /build/install-gems.sh

FROM --platform=linux/arm64 public.ecr.aws/sam/build-ruby4.0:1.174.1@sha256:21098addf33f0e9ef7e9e73f64afc1ca918be10e882486d94b9e8c2427426c9e AS ruby4.0-arm64
COPY layer/Gemfile layer/Gemfile.lock install-gems.sh /build/
RUN /build/install-gems.sh

FROM public.ecr.aws/sam/build-ruby3.3:1.174.1@sha256:05414c9464710198a1f15fa14bb4620a54003578aa06b6605c1f497c1f048511
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
