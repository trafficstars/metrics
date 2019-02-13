```
BenchmarkSortBuiltin-3               	 5000000	       103 ns/op	      32 B/op	       1 allocs/op
BenchmarkConsiderValueFlow-3         	 1000000	       527 ns/op	       0 B/op	       0 allocs/op
BenchmarkDoSliceFlow-3               	  300000	      1955 ns/op	       0 B/op	       0 allocs/op
BenchmarkGetPercentilesFlow-3        	 1000000	       523 ns/op	      96 B/op	       2 allocs/op
BenchmarkConsiderValueShortBuf-3     	 1000000	       328 ns/op	       0 B/op	       0 allocs/op
BenchmarkDoSliceShortBuf-3           	  300000	      1502 ns/op	       1 B/op	       0 allocs/op
BenchmarkGetPercentilesShortBuf-3    	 1000000	       483 ns/op	     136 B/op	       7 allocs/op
BenchmarkNewGaugeFloat64-3           	   50000	      7305 ns/op	    1312 B/op	      17 allocs/op
BenchmarkList-3                      	      30	  15752464 ns/op	  989904 B/op	      24 allocs/op
BenchmarkGenerateStorageKey-3        	10000000	        52.0 ns/op	       0 B/op	       0 allocs/op
BenchmarkGet-3                       	 3000000	       124 ns/op	       0 B/op	       0 allocs/op
BenchmarkRegistry-3                  	 2000000	       181 ns/op	       0 B/op	       0 allocs/op
BenchmarkRegistryReal-3              	  500000	       733 ns/op	       0 B/op	       0 allocs/op
BenchmarkAddToRegistryReal-3         	  500000	       780 ns/op	       0 B/op	       0 allocs/op
BenchmarkRegistryRealReal_lazy-3     	  500000	      1036 ns/op	     376 B/op	       4 allocs/op
BenchmarkRegistryRealReal_normal-3   	  500000	       944 ns/op	      40 B/op	       2 allocs/op
BenchmarkTagsString-3                	 1000000	       504 ns/op	       0 B/op	       0 allocs/op
BenchmarkTagsFastString-3            	 2000000	       235 ns/op	       0 B/op	       0 allocs/op
BenchmarkNewTimingBuffered-3         	   10000	     30824 ns/op	   84763 B/op	      53 allocs/op
BenchmarkNewTimingFlow-3             	   30000	     14287 ns/op	    3712 B/op	     103 allocs/op
```

Critical functions are `Get` (see `RegistryReal`) and `ConsiderValue`.
