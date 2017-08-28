#pragma once

// Can't call function pointers from Go as of writing

typedef int (*char_func)(char*);
typedef int (*float_func)(double);

void call_char_func(char_func func, char *arg);
void call_float_func(float_func func, double arg);
