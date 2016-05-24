prepare:
	wget -O c4.c https://raw.githubusercontent.com/rswier/c4/master/c4.c
	clang -S -emit-llvm -o c4.ll c4.c
	opt -mem2reg -S -o c4.ll c4.ll
	ll2dot -f -img c4.ll
