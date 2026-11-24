FROM public.ecr.aws/sam/build-nodejs22.x:1.174.0

ADD . /workspace
WORKDIR /workspace/packages/layer

RUN mkdir -p /build && \
    npm ci && \
    npm run clean && \
    npm run compile && \
    npm run install-externals && \
    zip -r /build/opentelemetry-nodejs-layer.zip build/workspace

CMD cp /build/opentelemetry-nodejs-layer.zip /out/
