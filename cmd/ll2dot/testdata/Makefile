SRC = $(wildcard *.c)
LL  = $(SRC:.c=.ll)

all: $(LL)

%.ll: %.c
	clang -S -emit-llvm -o $@ $<
	opt -S --mem2reg -o $@ $@
	ll2dot -f -img $@
