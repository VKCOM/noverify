PROJECT_DIR=./
NOVERIFY_ARGS=./src
NOVERIFY_MASTER_EXE=$(PWD)/noverify_master/build/noverify
NOVERIFY_CURRENT_EXE=$(PWD)/../../../build/noverify

clone_master:
	if [ ! -d noverify_master ]; then mkdir noverify_master && git clone https://github.com/VKCOM/noverify ./noverify_master; else git pull; fi

build_master: clone_master
	cd ./noverify_master && make build

compare: build_master
	go build
	cd ../../../ && make build
	cd $(PROJECT_DIR) && $(NOVERIFY_CURRENT_EXE) check --output-json --output='$(PWD)/new.json' $(NOVERIFY_ARGS)
	cd $(PROJECT_DIR) && $(NOVERIFY_MASTER_EXE) check --output-json --output='$(PWD)/old.json' $(NOVERIFY_ARGS)
	./stat -new new.json --old old.json > reports.md
