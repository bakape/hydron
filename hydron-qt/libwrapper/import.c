#include "import.h"

void call_char_func(char_func func, char *arg)
{
	func(arg);
}

void call_float_func(float_func func, double arg)
{
	func(arg);
}
