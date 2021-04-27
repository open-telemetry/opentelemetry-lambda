ARG runtime=python3.8

FROM public.ecr.aws/sam/build-${runtime}

ADD . /workspace

WORKDIR /workspace

RUN mkdir -p /build && \
  python3 -m pip install -r otel_sdk/requirements.txt -t /build/python && \
  python3 -m pip install -r otel_sdk/requirements-nodeps.txt -t /build/tmp --no-deps && \
  cp -r /build/tmp/* /build/python/ && \
  rm -rf /build/tmp && \
  cp -r otel_sdk/* /build/python && \
  mv /build/python/otel-instrument /build/otel-instrument && \
  chmod 755 /build/otel-instrument && \
  rm -rf /build/python/boto* && \
  rm -rf /build/python/urllib3* && \
  cd /build && \
  zip -r layer.zip otel-instrument python

CMD cp /build/layer.zip /out/layer.zip
