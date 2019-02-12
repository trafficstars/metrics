```
BenchmarkSortBuiltin-3                  	 2000000	       102 ns/op	      32 B/op	       1 allocs/op
BenchmarkConsiderValueFlow-3            	  300000	       525 ns/op	       0 B/op	       0 allocs/op
BenchmarkDoSliceFlow-3                  	  100000	      1888 ns/op	       0 B/op	       0 allocs/op
BenchmarkGetPercentilesFlow-3           	  300000	       527 ns/op	      96 B/op	       2 allocs/op
BenchmarkConsiderValueShortBuf-3        	  300000	       386 ns/op	       0 B/op	       0 allocs/op
BenchmarkDoSliceShortBuf-3              	  100000	      1438 ns/op	       3 B/op	       0 allocs/op
BenchmarkGetPercentilesShortBuf-3       	  200000	       614 ns/op	     136 B/op	       7 allocs/op
BenchmarkNewGaugeFloat64-3              	   20000	      6892 ns/op	    1312 B/op	      17 allocs/op
BenchmarkList-3                         	      10	  16152433 ns/op	  989904 B/op	      24 allocs/op
BenchmarkGenerateStorageKey-3           	 2000000	        55.4 ns/op	       0 B/op	       0 allocs/op
BenchmarkGet-3                          	 1000000	       120 ns/op	       0 B/op	       0 allocs/op
BenchmarkRegistry-3                     	 1000000	       179 ns/op	       0 B/op	       0 allocs/op
BenchmarkRegistryReal-3                 	  200000	       613 ns/op	       0 B/op	       0 allocs/op
BenchmarkAddToRegistryReal-3            	  200000	       820 ns/op	       0 B/op	       0 allocs/op
BenchmarkRegistryReal_withHiddenTag-3   	  300000	       419 ns/op	       0 B/op	       0 allocs/op
BenchmarkRegistryReal_FastTags-3        	  300000	       405 ns/op	       0 B/op	       0 allocs/op
BenchmarkTagsString-3                   	  300000	       526 ns/op	       0 B/op	       0 allocs/op
BenchmarkTagsFastString-3               	 1000000	       276 ns/op	       0 B/op	       0 allocs/op
BenchmarkNewTimingBuffered-3            	   10000	     32046 ns/op	   84766 B/op	      53 allocs/op
BenchmarkNewTimingFlow-3                	   10000	     13031 ns/op	    3712 B/op	     103 allocs/op
```

Critical functions are `Get` (see `RegistryReal`) and `ConsiderValue`.
