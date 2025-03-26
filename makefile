DIRS = ./utilities ./eam/models ./eam/components ./eam/handlers ./eam/server ./apps/manager ./apps/pip ./apps/pap ./apps/pdp ./apps/fsc/plugin/generic

.PHONY: all oas $(DIRS)

all: oas $(DIRS)

oas:
	+$(MAKE) -C ./oas

$(DIRS):
	+$(MAKE) -C $@ test
