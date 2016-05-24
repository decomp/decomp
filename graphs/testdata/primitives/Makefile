DOT = $(wildcard *.dot)
PNG = $(DOT:.dot=.png)

all: $(PNG)

%.png: %.dot
	dot -Tpng -o $@ $<
