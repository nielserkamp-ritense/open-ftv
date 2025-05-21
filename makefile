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

.PHONY: all oas $(DIRS)

all: oas $(DIRS)

oas:
	+$(MAKE) -C ./oas

$(DIRS):
	+$(MAKE) -C $@ test
