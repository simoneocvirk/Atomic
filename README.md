# Atomic

**USED AS DEMONSTRATION OF IBM Db2:** https://github.com/IBM-Db2-Developer/Atomic-Music-Discovery-With-Db2/tree/master/atomic-go

Build:

```
clang FFTHandler.c -L/usr/local/lib -lfftw3 -lm -Ofast -shared -fPIC -o libffthandler.so
go build .
```
