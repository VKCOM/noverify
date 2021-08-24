WORKING_DIR=/Users/petrmakhnev/psalm/
COMMAND=--index-only-files="./stubs" ./src
NOVERIFY_MASTER_EXE=$(PWD)/noverify_master/build/noverify
NOVERIFY_CURRENT_EXE=$(PWD)/../../../build/noverify

clone_master:
	if [ ! -d noverify_master ]; then mkdir noverify_master && git clone https://github.com/VKCOM/noverify ./noverify_master; fi

build_master: clone_master
	cd ./noverify_master && make build

compare: build_master
	go build
	cd ../../../ && make build
	cd $(WORKING_DIR) && $(NOVERIFY_CURRENT_EXE) check --output-json --output='$(PWD)/new.json' $(COMMAND)
	cd $(WORKING_DIR) && $(NOVERIFY_MASTER_EXE) check --output-json --output='$(PWD)/old.json' $(COMMAND)
	./stat -m --new new.json --old old.json > reports.md
