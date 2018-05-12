dexcounter
==========

*For counting methods. For Dex files.*

## What?

Dexcounter is a CLI tool for checking how many methods a library
will add to your dex file if you include it. It is inspired by
the old [methodscount.com][1] service, and borrows some of its tricks.

## How?

Dexcounter is written in [Go][2], so you can `go get` it:

    go get -u github.com/dhleong/dexcounter

Then, just pass the gradle-style dependency string like so:

    dexcounter io.reactivex.rxjava2:rxkotlin:2.2.0

and off it goes! `dexcounter` fetches a gradle environment (from
[this repo][3]) to resolve the library's dependencies, uses
the `dx` tool to count the method references in all of them,
and adds them all together. Its output currently looks something
like this:

```
io.reactivex.rxjava2:rxkotlin:2.2.0 TOTALS:
 Methods: 17277
  Fields: 6299

Dependency                                    Methods  Fields
io.reactivex.rxjava2:rxkotlin:2.2.0               679     126
  io.reactivex.rxjava2:rxjava:2.1.6             10276    5410
  org.jetbrains.kotlin:kotlin-stdlib:1.1.60      6291     741
  org.reactivestreams:reactive-streams:1.0.1        7       0
  org.jetbrains:annotations:13.0                   24      22
```

[1]: http://www.methodscount.com/
[2]: https://golang.org/
[3]: https://github.com/dhleong/dexcounter/tree/master/gradle
