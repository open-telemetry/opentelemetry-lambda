ARG runtime=python3.14

FROM public.ecr.aws/sam/build-${runtime}

ADD . /workspace

WORKDIR /workspace

# Install uv
RUN curl -LsSf https://astral.sh/uv/install.sh | sh && \
  export PATH="/root/.local/bin:$PATH"

RUN mkdir -p /build && \
  /root/.local/bin/uv venv /build/.venv && \
  /root/.local/bin/uv pip install --python /build/.venv/bin/python '.[instrumentation]' --target /build/python && \
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
