FROM public.ecr.aws/sam/build-python3.14:latest@sha256:08e8b22da8e66ee265e38d8aadbe3e700eb3d3c58325b7ad12ef4113d1ba7487

ADD . /workspace

WORKDIR /workspace

RUN mkdir -p /build && \
  python3 -m pip install -r otel_sdk/requirements.txt -t /build/python && \
  python3 -m pip install -r otel_sdk/nodeps-requirements.txt -t /build/tmp --no-deps && \
  # We need to use a `/build/tmp/` folder otherwise the instrumentation packages
  # do not get fully downloaded to the `opentelemetry/instrumentation/` path.
  cp -r /build/tmp/* /build/python/ && \
  rm -rf /build/tmp && \
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
