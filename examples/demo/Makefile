# Description of files:
#
# * foo.c:                original C source file
# * foo.ll:               LLVM IR assembly (produced by Clang)
# * foo_graphs/main.dot:  control flow graph of "main" function in GraphViz DOT format
# * foo_graphs/main.json: control flow recovery information (produced by restructure)
# * foo_pre.go:           unpolished decompiled Go source code (before post-processing).
# * foo.go:               decompiled Go source code (post-processed using go-post).

# Install tools:
#
#    Clang and LLVM
#    go get github.com/decomp/decomp/cmd/...
#    go get github.com/decomp/exp/cmd/dot2png
#    go get github.com/jackspirou/grind (fork of rsc.io/grind)
#    go get github.com/mewkiz/cmd/sar
#    go get golang.org/x/tools/cmd/goimports

all: foo.go foo.ll

%.ll: %.c
	# * Step 0
	#    - Convert original C source code to LLVM IR.
	#
	#    - input:  foo.c
	#    - output: foo.ll
	clang -S -emit-llvm -o $@ $<
	opt -S --mem2reg -o $@ $@

foo.go: foo.ll
	# * Step 1
	#    - Produce control flow graphs for each function of the LLVM IR.
	#
	#    - input:  foo.ll
	#    - output: foo_graphs/main.dot
	ll2dot -f $<
	# * Step 2
	#    - Perform control flow analysis on the control flow graphs to recover
	#      the high-level control flow primitives of each function.
	#
	#    - input:  foo_graphs/main.dot
	#    - output: foo_graphs/main.json
	#    - output (optional): foo_graphs/main_NNNN{a,b}.dot (when the debug flag -steps is used)
	cd foo_graphs; restructure -indent -steps -o main.json main.dot
	# (optional) Convert DOT control flow graphs to PNG images.
	dot2png foo_graphs/*.dot
	# * Step 3
	#    - Decompile the LLVM IR assembly to unpolished Go source code, using
	#      control flow recovery information produced by restructure.
	#
	#    - input:  foo.ll
	#    - input:  foo_graphs/main.json
	#    - output: foo_pre.go
	ll2go $< > foo_pre.go
	# * Step 4
	#    - Post-process the decompiled Go source code.
	#
	#    - input:  foo.go (actually a copy of foo_pre.go, since rewrite happens inplace)
	#    - output: foo.go
	cp foo_pre.go $@
	go-post $@
	# Force run "localid" post-processing rule (disabled by default).
	go-post -force localid $@
	# Move variable declarations closer to their use.
	grind $@
	# The rewrites of Grind enabled further rewrite rules of go-post.
	#
	# Note: use -diff to see the change of an given go-post rewrite; e.g.
	#
	#    go-post -diff -r forloop foo.go
	go-post -r forloop $@
	go-post -r unusedvar $@
	# Use sar to remove useless newlines.
	sar -i "\n\n" "\n" $@
	sar -i '([*<])\n' '$$1 ' $@
	# Use goimports to format the output Go source code.
	#
	#    - input:  foo.go
	#    - output: foo.go
	goimports -w $@
	# Add whitespace between top-level function declarations.
	sar -i "}\nfunc" "}\n\nfunc" $@

.PHONY: all clean

clean:
	$(RM) foo.ll foo_pre.go foo.go
	$(RM) -r foo_graphs
