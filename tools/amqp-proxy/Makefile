LIBS=`pkg-config libqpid-proton --libs` -lm
CFLAGS=`pkg-config libqpid-proton --cflags`

all: proxy

proxy: proxy.c proton-common.c
	gcc -g -Wall -o $@ $(LIBS) $(CFLAGS) $^

clean:
	rm -f *~ proxy
