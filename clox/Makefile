CC=gcc
CFLAGS= -Wall -Wextra

SRCS=$(wildcard *.c)
OBJS=$(patsubst %.c, build/%.o, $(SRCS))
EXEC=clox
TARGET=build/$(EXEC)

all: $(TARGET)
	@echo "////////// BUILD SUCCESS //////////"

build/%.o: %.c | build
	$(CC) $(CFLAGS) -c $< -o $@

$(TARGET): $(OBJS)
	$(CC) $(OBJS) -o $(TARGET)

build:
	mkdir -p build

clean:
	rm -rf build

.PHONY: all clean
