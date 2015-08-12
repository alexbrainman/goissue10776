#include <stdio.h>

int alexvariable = 4;

int alexfunction()
{
	return alexvariable;
}

int main()
{
	printf("hello %d\n", alexfunction());
}