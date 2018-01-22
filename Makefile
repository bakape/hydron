.PHONY: link

all: libthumbnailer.a

libthumbnailer.a: $(addsuffix .o, $(basename $(wildcard *.cc)))
	gcc-ar rcs libthumbnailer.a *.o

%.o: %.cc
	g++ -c -o $@ $^ -Wall -Wextra -I../lib/x86_64-unknown-linux-gnu/include -L../lib/x86_64-unknown-linux-gnu/lib

clean:
	rm -f *.o libthumbnailer.a