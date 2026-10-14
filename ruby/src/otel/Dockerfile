# Builds one layer with gems for each Ruby version on both x86_64 and arm64.

FROM --platform=linux/amd64 public.ecr.aws/sam/build-ruby3.3:1.169.0@sha256:4cadacfbc59358083f8f3091e6e947e2517b1613b819107a38a7d8dd5e89eea2 AS ruby3.3-amd64
COPY layer/Gemfile layer/Gemfile.lock install-gems.sh /build/
RUN /build/install-gems.sh

FROM --platform=linux/arm64 public.ecr.aws/sam/build-ruby3.3:1.169.0@sha256:4cadacfbc59358083f8f3091e6e947e2517b1613b819107a38a7d8dd5e89eea2 AS ruby3.3-arm64
COPY layer/Gemfile layer/Gemfile.lock install-gems.sh /build/
RUN /build/install-gems.sh

FROM --platform=linux/amd64 public.ecr.aws/sam/build-ruby3.4:1.169.0@sha256:33ddc1b6bce2d34d9782910fdf6d4265516abec02628569f2c5ec9489ff99570 AS ruby3.4-amd64
COPY layer/Gemfile layer/Gemfile.lock install-gems.sh /build/
RUN /build/install-gems.sh

FROM --platform=linux/arm64 public.ecr.aws/sam/build-ruby3.4:1.169.0@sha256:33ddc1b6bce2d34d9782910fdf6d4265516abec02628569f2c5ec9489ff99570 AS ruby3.4-arm64
COPY layer/Gemfile layer/Gemfile.lock install-gems.sh /build/
RUN /build/install-gems.sh

FROM --platform=linux/amd64 public.ecr.aws/sam/build-ruby4.0:1.169.0@sha256:353e66efaf32d4aa10900de67062ebe01798106ab7be043a2e43af59817cd080 AS ruby4.0-amd64
COPY layer/Gemfile layer/Gemfile.lock install-gems.sh /build/
RUN /build/install-gems.sh

FROM --platform=linux/arm64 public.ecr.aws/sam/build-ruby4.0:1.169.0@sha256:353e66efaf32d4aa10900de67062ebe01798106ab7be043a2e43af59817cd080 AS ruby4.0-arm64
COPY layer/Gemfile layer/Gemfile.lock install-gems.sh /build/
RUN /build/install-gems.sh

FROM public.ecr.aws/sam/build-ruby3.3:1.169.0@sha256:4cadacfbc59358083f8f3091e6e947e2517b1613b819107a38a7d8dd5e89eea2
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
