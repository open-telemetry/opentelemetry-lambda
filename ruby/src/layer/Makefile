
build-OTelLayer:
	mkdir -p $(ARTIFACTS_DIR)/ruby/gems/3.2.0
	bundle config set --global silence_root_warning 1
	bundle config set --local path 'ruby'
	bundle install
	cp -r ruby/ruby/3.2.0/* $(ARTIFACTS_DIR)/ruby/gems/3.2.0
	cp otel-handler $(ARTIFACTS_DIR)/otel-handler
	cp wrapper.rb $(ARTIFACTS_DIR)/wrapper.rb
	rm -rf $(ARTIFACTS_DIR)/ruby/gems/3.2.0/cache
