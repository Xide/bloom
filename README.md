bloom
===================

`Standard`, `Scalable` & `partitionned` bloom filter implementations.
Maths for this project were coded using [this implementation](http://gsd.di.uminho.pt/members/cbm/ps/dbloom.pdf).

----------

Usage
-------------

#### Standard
###### Basic usage
```go
// Create a new Filter of 512 bytes using 5 different hash functions
bf := bloom.New(512, 5)

// Insert element in the filter
bf.Feed("An item")

// Query for an element membership
bf.Match("An item")
// true

bf.Match("Another item")
// false
```

###### Merge filters
```go
bf := bloom.New(512, 5)
oth := bloom.New(512, 5)

bf.Feed("foo")
oth.Feed("bar")

bf.Merge(oth)

bf.Match("foo") && bf.Match("bar")
/// true
```


###### Export filter
```go
// Export filter as []byte for exportation
bytes, _ := bf.ToJSON()

// Export directly to filesystem (as json)
err := bf.ToFile("file.json")

```

Filters can then be imported with :

```go
bf, err := bloom.FromJSON(bytes)


// NotImplemented yet

bf, err := bloom.FromFile("file.json")
```

###### Filter fill ratio
```go
// Average ratio of bits set to 1; count each bit of the underlying array
// Might cause slowdown if used too much
bf.FillRatio()

// Optimization of the function above, estimate instead of counting
bf.EstimateFillRatio()
```
Both functions return a `float64` between 0 and 1.
 Please keep in mind that `EstimateFillRatio`  yield an approximate ratio, if you need precision below ~0.1, consider using directly `FillRatio`.

#### Scalable
```go
bf := bloom.NewScalableDefault()
```
