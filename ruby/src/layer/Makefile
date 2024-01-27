
build-OTelLayer:
	mkdir -p ruby
	mkdir -p $(ARTIFACTS_DIR)/ruby/gems/3.2.0
	bundler install --path ruby
	cp -r ruby/ruby/3.2.0/* $(ARTIFACTS_DIR)/ruby/gems/3.2.0
	cp otel-handler $(ARTIFACTS_DIR)/otel-handler
	rm -rf $(ARTIFACTS_DIR)/ruby/gems/3.2.0/cache
