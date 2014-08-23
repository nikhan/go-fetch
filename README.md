go-fetch
=====

![](http://4.bp.blogspot.com/-s_I4bwEz5BM/U2TpMsooVHI/AAAAAAAAAoQ/xQ3_ucCa7aU/s1600/fetch-mean-girls-quote-so-fetch-thats-so-fetch-Favim.com-372000.gif)

A small library that affords the use of simple jq/javascript/python-style accessors on nested interface{}s. go-fetch is *not* a replacement for properly unmarshalling JSON into  approriate structs and is intended to be used in situations where embedded data-accessing/querying is needed.

For example, given a map with the following structure:

```
{
    "foo":{
        "bar":[1,2,3]
    }
}
```
the second element of `bar` can be accessed by:

```
result, err := Fetch.Fetch(".foo.bar[2]", obj)
```
All queries must start with `.`, as this is refers to the value that is passed to go-fetch. Making a query of simple `.` will return the entire value itself.

go-fetch supports bracket accessors for maps as well, so if you need to access a key that has characters (such as a `.`) that need to be avoided, you can do so:

```
result, err := Fetch.Fetch(`.["foo"].bar[2]`, obj)
```

`Fetch.Fetch()`  is a convenience function and if performance is a concern it is highly recommended that you parse your query ahead of time with `Fetch.Parse()` and follow up with `Fetch.Run()` instead. 

```
query, _ := Fetch.Parse(`.["stop.trying"].to[0].make.fetch.happen`)
for{
    select {
        case m := <-data:
            Fetch.Run(query, m)
...

```


```
BenchmarkFetch              200000          18448 ns/op
BenchmarkFetchParseOnce     10000000          177 ns/op
BenchmarkNoFetch            20000000          112 ns/op
```

The above benchmarks were ran on a 2010 Macbook Pro. `BenchmarkFetch` is running `Fetch.Fetch()`. You can see that parsing the query every time can be costly. The second benchmark, `BenchmarkFetchParseOnce` compiles the query once with `Fetch.Parse()`. Finally `BenchmarkNoFetch` is testing the time it takes to do all of the assertions and checking one would need to do when accessing a deeply nested value.
