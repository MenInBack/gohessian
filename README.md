# gohessian

A sample hessian protocol implementation for Go-lang


### Install
```sh
$ go get github.com/MenInBack/gohessian
```

### Usage

```go
package main

import (
    "fmt"
    gh "github.com/MenInBack/gohessian"
)

func main() {
    c := gh.NewClient("http://www.example.com", "/helloworld")
    res, err := c.Invoke("sendInt", 1)
    if err != nil {
        fmt.Printf("Hessian Invoke error:%s\n",err)
        return
    }
    fmt.Printf("Hessian Invoke Success, result:%s\n", res)
}
```
