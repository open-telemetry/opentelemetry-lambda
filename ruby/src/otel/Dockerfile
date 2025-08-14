FROM ubuntu:latest

# Comma-separated list of Ruby versions to build, e.g. "3.4.0" or "3.4.0,3.3.0".
ARG RUBY_VERSIONS="3.2.0,3.3.0,3.4.0"

RUN mkdir /build

COPY . /build

RUN apt-get update && apt-get install -y git curl libssl-dev libreadline-dev zlib1g-dev autoconf bison build-essential libyaml-dev libreadline-dev libncurses5-dev libffi-dev libgdbm-dev zip

RUN  git clone https://github.com/rbenv/rbenv.git ~/.rbenv \
     && git clone https://github.com/rbenv/ruby-build.git ~/.rbenv/plugins/ruby-build \
     && git clone https://github.com/rbenv/rbenv-default-gems.git ~/.rbenv/plugins/rbenv-default-gems \
     && echo 'export PATH="$HOME/.rbenv/bin:$PATH"' >> ~/.profile \
     && echo 'eval "$(rbenv init -)"' >> ~/.profile \
     && echo 'eval "$(rbenv init -)"' >> ~/.bashrc \
     && echo 'bundler' > ~/.rbenv/default-gems

RUN echo 'alias be="bundle exec"' >> ~/.bashrc
RUN echo 'alias be="bundle exec"' >> ~/.profile

# install rubies to build our gem against Gemfile
RUN set -e; . ~/.profile; \
    cd /root/.rbenv/plugins/ruby-build && git pull && cd -; \
    for v in $(echo "$RUBY_VERSIONS" | tr ',' ' '); do \
      echo "Installing Ruby $v"; \
      rbenv install -s "$v"; \
    done

WORKDIR /build/layer

RUN set -e; . ~/.profile; \
    for v in $(echo "$RUBY_VERSIONS" | tr ',' ' '); do \
      echo "Bundler install for Ruby $v"; \
      rbenv local "$v"; \
      bundle install; \
    done

RUN set -e; . ~/.profile; \
    for v in $(echo "$RUBY_VERSIONS" | tr ',' ' '); do \
      cd "/root/.rbenv/versions/$v/lib/ruby/gems/"; \
      # Determine the RubyGems ABI dir (e.g., 3.4.0) for this Ruby $v
      abi_dir=$(RBENV_VERSION="$v" ruby -e 'print RbConfig::CONFIG["ruby_version"]' 2>/dev/null || true); \
      # Fallback: find a directory matching major.minor.* when patch-level differs (e.g., 3.4.4 -> 3.4.0)
      if [ -z "$abi_dir" ] || [ ! -d "$abi_dir" ]; then \
        major_minor=$(echo "$v" | cut -d. -f1-2); \
        abi_dir=$(find . -maxdepth 1 -type d -name "${major_minor}.*" -printf "%f" -quit) || true; \
      fi; \
      if [ -z "$abi_dir" ] || [ ! -d "$abi_dir" ]; then \
        echo "Could not locate RubyGems dir for Ruby $v under $(pwd)" >&2; exit 1; \
      fi; \
      zip -r "gems-$v.zip" "$abi_dir"/; \
    done

RUN mkdir -p /build/ruby/gems
WORKDIR /build/ruby/gems
RUN set -e; for v in $(echo "$RUBY_VERSIONS" | tr ',' ' '); do \
      cp "/root/.rbenv/versions/$v/lib/ruby/gems/gems-$v.zip" .; \
      unzip -q "gems-$v.zip" && rm "gems-$v.zip"; \
    done \
    && ls -al /build/ruby/gems

# rm gem cache
RUN set -e; for v in $(echo "$RUBY_VERSIONS" | tr ',' ' '); do \
      rm -rf "/root/.rbenv/versions/$v/lib/ruby/gems/$v/cache"/* || true; \
    done

# zip all the gems
WORKDIR /build
RUN cp layer/otel-handler . && cp layer/wrapper.rb .
RUN zip -r opentelemetry-ruby-layer.zip ruby/ otel-handler wrapper.rb

CMD cp /build/opentelemetry-ruby-layer.zip /out/opentelemetry-ruby-layer.zip
