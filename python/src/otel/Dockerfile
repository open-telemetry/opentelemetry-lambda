FROM public.ecr.aws/sam/build-python3.14:latest@sha256:c1979e487cdfc03dc27901b5a6b2c0d105ee63e7ed15fc4450d3b3c4248fadd6

ADD . /workspace

WORKDIR /workspace

RUN mkdir -p /build && \
  python3 -m pip install -r otel_sdk/requirements.txt -t /build/python && \
  mv otel_sdk/otel_wrapper.py /build/python && \
  mv otel_sdk/otel-instrument /build && \
  mv otel_sdk/otel-handler /build && \
  chmod 755 /build/otel-instrument && \
  chmod 755 /build/otel-handler && \
  rm -rf /build/python/boto* && \
  rm -rf /build/python/urllib3* && \
  cd /build && \
  zip -r opentelemetry-python-layer.zip otel-handler otel-instrument python

CMD cp /build/opentelemetry-python-layer.zip /out/opentelemetry-python-layer.zip
