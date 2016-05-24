int main(int argc, char **argv) {
	int x;

	x = 10;
	switch (x) {
	case 3:
		x = 33;
	case 6:
		x = 66;
	case 9:
		x = 99;
	default:
		x = 42;
	}
	return x;
}
