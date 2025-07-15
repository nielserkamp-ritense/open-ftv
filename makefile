DIRS := \
 ./utilities \
 ./eam/authentication \
 ./eam/authorization \
 ./eam/models \
 ./eam/config \
 ./eam/log \
 ./eam/mapping \
 ./eam/mimetype \
 ./eam/handlers \
 ./eam/server \
 ./eam/pep \
 ./eam/pip \
 ./eam/pap \
 ./eam/pdp/controller \
 ./eam/pdp/cedar-embedded \
 ./eam/pdp/cerbos-api \
 ./eam/pdp/odrl \
 ./eam/pdp/opa-embedded \
 ./eam/pdp/openfga-embedded \
 ./eam/pdp/xacml \
 ./mock/datasources/data \
 ./apps/fsc-auth \
 ./apps/pdp \
 ./apps/pip \
 ./apps/pap \
 ./apps/manager

.PHONY: all
all: oas $(DIRS) vlierdam rdw rvig

.PHONY: e2e
e2e: vlierdam rdw rvig

.PHONY: oas
oas:
	+$(MAKE) -C ./oas

.PHONY: $(DIRS)
$(DIRS):
	+$(MAKE) -C $@ test

.PHONY: vlierdam
vlierdam:
	@./e2e/gemeente-vlierdam/test.sh

.PHONY: rdw
rdw:
	@./e2e/rdw/test.sh

.PHONY: rvig
rvig:
	@./e2e/rvig/test.sh
