NAME := tahmkench
REGISTRY_PREFIX :=

.PHONY: build
build:
ifneq ($(name),)
	docker build -t $(REGISTRY_PREFIX)$(NAME):$(name) .
else
	docker build -t $(REGISTRY_PREFIX)$(NAME):latest .
endif

.PHONY: push
push:
ifneq ($(name),)
	docker push $(REGISTRY_PREFIX)$(NAME):$(name)
else
	docker push $(REGISTRY_PREFIX)$(NAME):latest
endif
