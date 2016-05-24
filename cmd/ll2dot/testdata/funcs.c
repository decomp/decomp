int foo(int a, int b) {
	int x;

	if (b < 3) {
		b <<= 10;
	}
	for (x = 0; a < b; a++) {
		x += a;
	}
	return x;
}

int bar(int x) {
	while (x < 1000) {
		x *= 2;
	}
	return x;
}

int main(int argc, char **argv) {
	int x;
	if (x < 3) {
		return bar(x);
	}
	return bar(x*2);
}
