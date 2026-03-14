BINARY  := bootstrap
OUTDIR  := dist

.PHONY: build clean test

build:
	@mkdir -p $(OUTDIR)
	go build -o $(OUTDIR)/$(BINARY) .

clean:
	rm -rf $(OUTDIR)

test:
	bash tests/test.sh
