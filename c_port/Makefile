CC=gcc
CFLAGS=-O2 -Wall -I.
LDFLAGS=-lcurl

all: radp asnmap

cJSON.o: cJSON.c cJSON.h
	$(CC) $(CFLAGS) -c cJSON.c

radp: radp.c cJSON.o
	$(CC) $(CFLAGS) radp.c cJSON.o -o radp $(LDFLAGS)

asnmap: asnmap.c cJSON.o
	$(CC) $(CFLAGS) asnmap.c cJSON.o -o asnmap $(LDFLAGS)

clean:
	rm -f *.o radp asnmap
