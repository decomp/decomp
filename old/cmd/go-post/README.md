# go-post

`go-post` is a tool which post-processes Go source code to make it more idiomatic.

The development of `go-post` started with an unmodified copy of [gofix](https://golang.org/cmd/fix/) from [golang/go@edcad86](https://github.com/golang/go/commit/edcad8639a902741dc49f77d000ed62b0cc6956f).

## Installation

```shell
go get github.com/decomp/decomp/cmd/go-post
```

## Usage

```
Usage: go-post [-diff] [-r fixname,...] [-force fixname,...] [path ...]

Flags:
  -diff
        display diffs instead of rewriting files
  -force string
        force these fixes to run even if the code looks updated
  -r string
        restrict the rewrites to this comma-separated list

Available rewrites are:

assignbinop
    Replace "x = x + z" with "x += z".

deadassign
    Remove "x = x" assignments.

forloop
    Add initialization and post-statements to for-loops.

localid
    Replace the use of local variable IDs with their definition.

mainret
    Replace return statements with calls to os.Exit in the "main" function.

unresolved
    Replace assignment statements with declare and initialize statements at the first occurance of an unresolved identifier.
```

## Examples

### all

```bash
$ go-post foo.go
```

INPUT: [foo_orig.go](examples/foo_orig.go)

```go
package main

func main() {
    i = 0
    x = 0
    for i < 10 {
        _4 := x < 100
        x = x
        if _4 {
            _6 := 3 * i
            _7 := x + _6
            x = _7
        }
        _10 := i + 1
        i = _10
        x = x
    }
    return x
}
```

OUTPUT: [foo.go](examples/foo.go)

```go
package main

import "os"

func main() {
    x := 0
    for i := 0; i < 10; i++ {
        if x < 100 {
            x += 3 * i
        }
    }
    os.Exit(x)
}
```

### unresolved

```bash
$ go-post -diff -r "unresolved" foo.go
```

```patch
diff foo.go fixed/foo.go
--- /tmp/go-fix921106208    2015-03-27 00:58:58.757731358 +0100
+++ /tmp/go-fix992885759    2015-03-27 00:58:58.757731358 +0100
@@ -1,8 +1,8 @@
 package main
 
 func main() {
-   i = 0
-   x = 0
+   i := 0
+   x := 0
    for i < 10 {
        _4 := x < 100
        x = x

```

### mainret

```bash
$ go-post -diff -r "mainret" foo.go
```

```patch
diff foo.go fixed/foo.go
--- /tmp/go-fix248843565    2015-03-27 01:02:07.850040290 +0100
+++ /tmp/go-fix745944744    2015-03-27 01:02:07.850040290 +0100
@@ -1,5 +1,7 @@
 package main
 
+import "os"
+
 func main() {
    i := 0
    x := 0
@@ -15,5 +17,5 @@
        i = _10
        x = x
    }
-   return x
+   os.Exit(x)
 }
```

### localid

```bash
$ go-post -diff -r="localid" foo.go
```

```patch
diff foo.go fixed/foo.go
--- /tmp/go-fix510478144    2015-03-27 01:04:31.052597829 +0100
+++ /tmp/go-fix575640991    2015-03-27 01:04:31.052597829 +0100
@@ -6,15 +6,17 @@
    i := 0
    x := 0
    for i < 10 {
-       _4 := x < 100
+
        x = x
-       if _4 {
-           _6 := 3 * i
-           _7 := x + _6
-           x = _7
+       if x < 100 {
+
+           x = x +
+               3*i
+
        }
-       _10 := i + 1
-       i = _10
+
+       i = i + 1
+
        x = x
    }
    os.Exit(x)
```

### assignbinop

```bash
go-post -diff -r="assignbinop" foo.go
```

```patch
diff foo.go fixed/foo.go
--- /tmp/go-fix616309309    2015-03-27 01:09:50.707532776 +0100
+++ /tmp/go-fix946417784    2015-03-27 01:09:50.707532776 +0100
@@ -8,10 +8,9 @@
    for i < 10 {
        x = x
        if x < 100 {
-           x = x +
-               3*i
+           x += 3 * i
        }
-       i = i + 1
+       i++
        x = x
    }
    os.Exit(x)
```

### deadassign

```bash
$ go-post -diff -r="deadassign" foo.go
```

```patch
diff foo.go fixed/foo.go
--- /tmp/go-fix088382919    2015-03-27 01:11:46.960236315 +0100
+++ /tmp/go-fix911863418    2015-03-27 01:11:46.960236315 +0100
@@ -6,12 +6,12 @@
    i := 0
    x := 0
    for i < 10 {
-       x = x
+
        if x < 100 {
            x += 3 * i
        }
        i++
-       x = x
+
    }
    os.Exit(x)
 }
```

### forloop

```bash
$ go-post -diff -r="forloop" foo.go
```

```patch
diff foo.go fixed/foo.go
--- /tmp/go-fix195227136    2015-03-27 01:13:28.103021710 +0100
+++ /tmp/go-fix984907103    2015-03-27 01:13:28.103021710 +0100
@@ -3,13 +3,13 @@
 import "os"
 
 func main() {
-   i := 0
+
    x := 0
-   for i < 10 {
+   for i := 0; i < 10; i++ {
        if x < 100 {
            x += 3 * i
        }
-       i++
+
    }
    os.Exit(x)
 }
```

## Public domain

The source code and any original content of this repository is hereby released into the [public domain].

[public domain]: https://creativecommons.org/publicdomain/zero/1.0/

## License

Any code or documentation directly derived from the [standard Go source code](https://github.com/golang/go) is governed by a [BSD license](http://golang.org/LICENSE).
