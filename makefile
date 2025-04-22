DIRS := \
 ./utilities \
 ./eam/authentication \
 ./eam/authorization \
 ./eam/config \
 ./eam/handlers \
 ./eam/log \
 ./eam/mapping \
 ./eam/mimetype \
 ./eam/models \
 ./eam/pap \
 ./eam/pdp/controller \
 ./eam/pdp/cedar-embedded \
 ./eam/pdp/cerbos-api \
 ./eam/pdp/odrl \
 ./eam/pdp/opa-embedded \
 ./eam/pdp/openfga-embedded \
 ./eam/pdp/xacml \
 ./eam/pep \
 ./eam/pip \
 ./eam/server \
 ./apps/fsc-auth \
 ./apps/manager \
 ./apps/pap \
 ./apps/pdp \
 ./apps/pip

.PHONY: all oas $(DIRS)

all: oas $(DIRS)

oas:
	+$(MAKE) -C ./oas

$(DIRS):
	+$(MAKE) -C $@ test
