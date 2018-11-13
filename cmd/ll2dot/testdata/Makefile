C_SRC=$(wildcard *.c)
LL_SRC=$(C_SRC:.c=.ll)
GRAPH_DIRS=$(LL_SRC:.ll=_graphs)

all: $(GRAPH_DIRS)

%_graphs: %.ll
	ll2dot -f -img $<

prepare: $(LL_SRC)

%.ll: %.c
	clang -S -emit-llvm -o $@ $<
	opt -S --mem2reg -o $@ $@

clean:
	rm -f *.ll
	rm -rf *_graphs

.PHONY: all prepare clean
