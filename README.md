# Overview
A minimal Go implementation of [RFC 959](https://www.rfc-editor.org/rfc/rfc959).

Project mainly exists to practice/use/study the interesting parts of Go (server-side network programming). 

```
   5.1.  MINIMUM IMPLEMENTATION

      In order to make FTP workable without needless error messages, the
      following minimum implementation is required for all servers:

         TYPE - ASCII Non-print
         MODE - Stream
         STRUCTURE - File, Record
         COMMANDS - USER, QUIT, PORT,
                    TYPE, MODE, STRU,
                      for the default values
                    RETR, STOR,
                    NOOP.

      The default values for transfer parameters are:

         TYPE - ASCII Non-print
         MODE - Stream
         STRU - File

      All hosts must accept the above as the standard defaults.
```

# How to use
```bash

# in a set up shell, cd to project root
mkdir temp

echo "hello world!" > ./temp/hello.txt

go run ./cmd/main.go

# in another shell
source ./scripts/curl.sh
```
