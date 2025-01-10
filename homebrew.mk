##########################################################################
#
# make tasks for generating a new `bbctl.rb` homebrew formula
#
##########################################################################

BREWER_ROOT := ./scripts/brewer

.PHONY: brew-regenerate brewer-test

brew-regenerate:
	@# todo: bbctl CI job needs permissions to push this as a new commit to the homebrew repo and open an MR
	@go run $(BREWER_ROOT)

brewer-test:
	go test -v $(BREWER_ROOT)/... --coverprofile=$(BREWER_ROOT)/cover.txt
	go tool cover -html=$(BREWER_ROOT)/cover.txt