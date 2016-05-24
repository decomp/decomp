int main(int argc, char **argv) {
	int i, x;

	x = 0;
	for (i = 0; i < 10; i++) {
		if (x < 100) {
			x += 3*i;
		}
	}
	return x;
}
