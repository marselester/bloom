# Bloom Filter

[![Documentation](https://godoc.org/github.com/marselester/bloom?status.svg)](https://godoc.org/github.com/marselester/bloom)
[![Go Report Card](https://goreportcard.com/badge/github.com/marselester/bloom)](https://goreportcard.com/report/github.com/marselester/bloom)

This is a Bloom filter implementation just for fun. Databases use them to reduce the disk lookups for non-existent rows or columns.
Avoiding costly disk lookups considerably increases the performance of a database query operation.

A [Bloom filter](https://en.wikipedia.org/wiki/Bloom_filter) is a space-efficient probabilistic data structure
that is used to test whether an element is a member of a set.
False positive matches are possible, but false negatives are not – in other words, a query returns either "possibly in set" or
"definitely not in set". Elements can be added to the set, but not removed; the more elements that are added to the set,
the larger the probability of false positives.

A Bloom filter of a fixed size can represent a set with an arbitrarily large number of elements (4,294,967,295 in this implementation); adding an element never fails due to the data structure "filling up".

## Usage Example

For more details please refer to [documentation](https://godoc.org/github.com/marselester/bloom).

```go
package main

import (
    "fmt"
    "log"

    "github.com/marselester/bloom"
)

func main() {
    const maxEmails = 100
    const prob = 0.01
    bf, err := bloom.New(maxEmails, prob)
    if err != nil {
        log.Fatalf("Bloom filter is not created: %v", err)
    }

    email := []byte("alice@example.com")
    // Must operations panic on err caused by hash function, e.g., failed to convert hex to decimal
    // (that shouldn't happen normally). You can use Add/Has which return errors.
    bf.MustAdd(email)
    if bf.MustHave(email) {
        fmt.Print("Alice's email possibly is in the set.")
    } else {
        fmt.Print("Alice's email definitely is not in the set.")
    }
}
```

## Algorithm

The idea is to "convert" an element into several bit array's indexes ("coordinates" or positions).
For example, `"test"` string is transformed into `7, 36, 32, 37` offsets of a bit array.
To add an element, set bits to 1 on those positions.
To check if an element is a member of a set, all bits must be 1 on those positions.

```
 1  1  0  0  0  1  0  0  0  0  0  0  0  0  0  0  0  0  0  0  0  0  0  0  0  0  0 0 0 0 1 0 0 0 0 0 0
36 35 34 33 32 31 30 29 28 27 26 25 24 23 22 21 20 19 18 17 16 15 14 13 12 11 10 9 8 7 6 5 4 3 2 1 0
```

The element was transformed into indexes by applying 4 hash functions. A hash function (in this package)
uses sha256 digest (hex) and converts it into a decimal number (from base 16 number into base 10 number).
Since we need 4 distinct hash functions, we can append a number to an element.

```
hex2dec(sha256("test0")) == 7
hex2dec(sha256("test1")) == 36
hex2dec(sha256("test2")) == 32
hex2dec(sha256("test3")) == 37
```

Based on desired probability of an error (false positives) and number of elements you intend to add,
it's possible to calculate optimal number of hash functions and length of a bit array.
For example, 1,000,000 elements set with 0.01 error rate requires 9,585,059 bits (1.198 MB) of storage.

```
max_elements = 1000000
desired_prob = 0.01
hash_qty = -ln(desired_prob) / ln(2)
bit_array_len = -max_elements * ln(desired_prob) / (ln(2) * ln(2))
```

## Tests

Install go-fuzz.

```sh
$ go get github.com/dvyukov/go-fuzz/go-fuzz
$ go get github.com/dvyukov/go-fuzz/go-fuzz-build
```

Start the fuzzing and see if there are crashers.

```sh
$ go-fuzz-build github.com/marselester/bloom
$ go-fuzz -bin=bloom-fuzz.zip -workdir=fuzz
```

## Benchmarks

This Bloom filter implementation certainly has room for improvement, e.g., reduce the number of memory allocations
or try to hash in parallel. Though the objective here is to keep code simple.

```sh
$ go test -bench=. -benchmem
BenchmarkFilter_Add/1.198MB-4         	  300000	      4254 ns/op	    1888 B/op	      30 allocs/op
BenchmarkFilter_Add/2.573GB-4         	  300000	      5198 ns/op	    1888 B/op	      30 allocs/op
BenchmarkFilter_Has/1.198MB-4         	  300000	      4269 ns/op	    1888 B/op	      30 allocs/op
BenchmarkFilter_Has/2.573GB-4         	  300000	      4285 ns/op	    1888 B/op	      30 allocs/op
```
