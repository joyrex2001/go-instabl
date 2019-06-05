# go-instabl

go-instabl is a static analysis tool that will calculate the instability
metric. The closer the value gets to 0, the more packages inside the code base
rely on this package, and the more responsible the package is. This means, the
package is considered stable, and changes to this package may have significant
impact on other packages. When the value is closer to 1, the package is more
resilient to changes, as it is less coupled to other packages.

Install with:
```
go get github.com/joyrex2001/go-instabl
```

And check analyze your repo with:
```
go-instabl src/github.com/meand/myrepo
```

See also: https://en.wikipedia.org/wiki/Software_package_metrics
