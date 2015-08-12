package main

/*
#include<stdio.h>
int getint() {
	printf("hello\n");
	return 3;
}
*/
import "C"

var alexval = "alex"

func alexfunc(s string) string {
	return s
}

func main() {
	println(alexfunc(alexval), C.getint())
}
