```
goos: linux
goarch: amd64
pkg: github.com/trafficstars/fastmetrics
BenchmarkSortBuiltin-3                  	 2000000	        99.8 ns/op	      32 B/op	       1 allocs/op
BenchmarkConsiderValueFlow-3            	  300000	       522 ns/op	       0 B/op	       0 allocs/op
BenchmarkDoSliceFlow-3                  	  100000	      1840 ns/op	       0 B/op	       0 allocs/op
BenchmarkGetPercentilesFlow-3           	  300000	       541 ns/op	      96 B/op	       2 allocs/op
BenchmarkConsiderValueShortBuf-3        	  500000	       354 ns/op	       0 B/op	       0 allocs/op
BenchmarkDoSliceShortBuf-3              	  100000	      1539 ns/op	       3 B/op	       0 allocs/op
BenchmarkGetPercentilesShortBuf-3       	  200000	       565 ns/op	     136 B/op	       7 allocs/op
BenchmarkNewGaugeFloat64-3              	   20000	      7231 ns/op	    1312 B/op	      17 allocs/op
BenchmarkList-3                         	      10	  15451234 ns/op	  989904 B/op	      24 allocs/op
BenchmarkGenerateStorageKey-3           	 1000000	       157 ns/op	       0 B/op	       0 allocs/op
BenchmarkGet-3                          	 1000000	       223 ns/op	       0 B/op	       0 allocs/op
BenchmarkRegistry-3                     	  500000	       258 ns/op	       0 B/op	       0 allocs/op
BenchmarkRegistryReal-3                 	  200000	       800 ns/op	       0 B/op	       0 allocs/op
BenchmarkAddToRegistryReal-3            	  200000	       866 ns/op	       0 B/op	       0 allocs/op
BenchmarkRegistryReal_withHiddenTag-3   	  300000	       612 ns/op	       0 B/op	       0 allocs/op
BenchmarkRegistryReal_FastTags-3        	  300000	       644 ns/op	       0 B/op	       0 allocs/op
BenchmarkTagsString-3                   	  200000	       648 ns/op	       0 B/op	       0 allocs/op
BenchmarkTagsFastString-3               	  300000	       472 ns/op	       0 B/op	       0 allocs/op
BenchmarkNewTimingBuffered-3            	    5000	     50546 ns/op	   84771 B/op	      53 allocs/op
BenchmarkNewTimingFlow-3                	   10000	     13747 ns/op	    3712 B/op	     103 allocs/op
PASS
ok  	github.com/trafficstars/fastmetrics	25.890s
```

Critical functions are `Get` (see `RegistryReal`) and `ConsiderValue`.
