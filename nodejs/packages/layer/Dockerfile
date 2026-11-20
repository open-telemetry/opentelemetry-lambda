FROM public.ecr.aws/sam/build-nodejs22.x:1.174.0

ADD . /workspace
WORKDIR /workspace

RUN npm install && \
    npm run clean && \
    npm run compile && \
    npm run install-externals && \
    cd build/workspace && \
    zip -r /build/opentelemetry-nodejs-layer.zip .

CMD cp /build/opentelemetry-nodejs-layer.zip /out/
