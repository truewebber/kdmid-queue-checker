# helm vars
HELM_CHART_PATH = ./helm/kdmid
RELEASE_NAME = kdmid
PACKAGE_VERSION = 0.1.0
PACKAGE_DESTINATION = .
APP_VERSION := $(shell git rev-parse HEAD)

# app vars
include .env

# Default target
all: upgrade clean

# Target to upgrade or install the helm release
upgrade:
	@CHART_NAME=$$(helm show chart $(HELM_CHART_PATH) | grep 'name:' | awk -F ': ' '{print $$2}') && \
	echo "$$CHART_NAME, Version: $(PACKAGE_VERSION), AppVersion: $(APP_VERSION)" && \
	helm package $(HELM_CHART_PATH) --version $(PACKAGE_VERSION) --app-version $(APP_VERSION) --destination $(PACKAGE_DESTINATION) && \
	CHART="$(PACKAGE_DESTINATION)/$$CHART_NAME-$(PACKAGE_VERSION).tgz" && \
	helm upgrade --install $(RELEASE_NAME) $$CHART \
	--set-string app.two_captcha_api_key=$(TWO_CAPTCHA_API_KEY) \
	--set-string app.artifacts_directory=$(ARTIFACTS_DIRECTORY) \
	--set-string app.recipient_storage_directory=$(RECIPIENT_STORAGE_DIRECTORY) \
	--set-string app.recipient_storage_limit=$(RECIPIENT_STORAGE_LIMIT) \
	--set-string app.telegram_bot_token=$(TELEGRAM_BOT_TOKEN) \
	--set-string app.proxy_url=$(PROXY_URL) \
	--namespace apps --wait;

# Optional clean target
clean:
	@rm $(PACKAGE_DESTINATION)/*-$(PACKAGE_VERSION).tgz

# Help target
help:
	@echo "Available targets:"
	@echo "  make upgrade   - Upgrade or install helm release"
	@echo "  make clean     - Clean up resources"
	@echo "  make help      - Display this help message"
