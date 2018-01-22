.PHONY: link

all: link

link: $(addsuffix .o, $(basename $(wildcard *.cc)))
	g++ -o linked *.o

%.bc: %.cc
	g++ $^ -o $@ -Wall -Wextra

clean:
	rm -f *.bc