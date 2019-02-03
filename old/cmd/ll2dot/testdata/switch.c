int main(int argc, char **argv) {
	int x;

	switch (argc) {
	case 2:
		x = 22;
		break;
	case 3:
		x = 33;
		break;
	case 5:
		x = 55;
		break;
	case 7:
		x = 77;
		break;
	default:
		x = 42;
		break;
	}
	return x;
}
