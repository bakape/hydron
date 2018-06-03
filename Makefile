.PHONY: link

lib_dir=./lib/x86_64-unknown-linux-gnu/lib

all: libthumbnailer.a

libthumbnailer.a: $(addsuffix .o, $(basename $(wildcard *.cc)))
	gcc-ar rcs libthumbnailer.a $(lib_dir)/libswscale.a $(lib_dir)/libavformat.a $(lib_dir)/libavutil.a  *.o

%.o: %.cc
	g++ -flto -O3 -c -o $@ $^ -Wall -Wextra -Wno-switch -I./lib/x86_64-unknown-linux-gnu/include -L$(lib_dir)

clean:
	rm -f *.o libthumbnailer.a
