C_SRC=$(wildcard *.c)
LL_SRC=$(C_SRC:.c=.ll)
GO_SRC=$(LL_SRC:.ll=.go)

all: $(GO_SRC) $(LL_SRC)

%.go: $(LL_SRC)
	ll2go -f $<

%.ll: %.c
	clang -S -emit-llvm -o $@ $<
	opt -S --mem2reg -o $@ $@

clean:
	rm -rf *.go *.ll *_graphs/
