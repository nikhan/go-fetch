go-fetch
=====

A small library that affords the use of simple jq/javascript/python-style accessors on nested interface{}s. go-fetch is *not* a replacement for properly unmarshalling JSON into appropriate structs and is intended to be used in situations where embedded data-accessing/querying is needed.

Documentation available at http://godoc.org/github.com/nikhan/go-fetch.

For example, given a map with the following structure:

```
{
    "foo":{
        "bar":[1,2,3]
    }
}
```
the third element of `bar` can be accessed by:

```
result, err := Fetch.Fetch(".foo.bar[2]", obj)
```
All queries must start with `.`, as this refers to the root of the value that is passed to go-fetch. Making a query `.` will return the entire value itself.

go-fetch supports bracket accessors for maps, so if you need to access a key that has characters that need to be avoided (such as a `.`,`#`,`$`,`*`,`%`,`!`), you can do so:

```
result, err := Fetch.Fetch(`.["foo"].bar[2]`, obj)
```

`Fetch.Fetch()`  is a convenience function that runs both `Fetch.Parse()` and `Fetch.Run()`. If you have a situation where you will be running the same query over lots of values it is highly recommended that you `Fetch.Parse()` your query once and `Fetch.Run()` each value that needs to be queried. 

```
query, _ := Fetch.Parse(`.["stop.trying"].to[0].make.fetch.happen`)
for{
    select {
        case m := <-data:
            Fetch.Run(query, m)
...

```


```
BenchmarkFetch	  				200000	     17778 ns/op
BenchmarkFetchParseOnce	  	  10000000	       168 ns/op
BenchmarkNoFetch			  20000000	       117 ns/op
BenchmarkNoFetchNoCheck		2000000000	      1.45 ns/op
```

The above benchmarks were run on a 2010 Macbook Pro. `BenchmarkFetch` is running `Fetch.Fetch()`. You can see that parsing the query every time can be costly. The second benchmark, `BenchmarkFetchParseOnce` compiles the query once with `Fetch.Parse()`.  `BenchmarkNoFetch` is testing the time it takes to do all of the assertions on a map of interfaces{}. Finally, `BenchmarkNoFetchNoCheck` is what happens when dealing with properly typed structs.
