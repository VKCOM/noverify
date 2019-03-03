In order to compile and run the demo, do the following:

```
$ cp $(go env GOROOT)/misc/wasm/wasm_exec.js .
$ env GOOS=js GOARCH=wasm go build -o main.wasm
$ go get -u github.com/shurcooL/goexec
$ goexec 'http.ListenAndServe(":8080", http.FileServer(http.Dir(".")))'
```

Then open http://localhost:8080/ in your browser and enjoy the demo!
