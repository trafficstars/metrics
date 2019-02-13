```
BenchmarkSortBuiltin-3              	 2000000	        99.0 ns/op	      32 B/op	       1 allocs/op
BenchmarkConsiderValueFlow-3        	  300000	       508 ns/op	       0 B/op	       0 allocs/op
BenchmarkDoSliceFlow-3              	  100000	      1925 ns/op	       0 B/op	       0 allocs/op
BenchmarkGetPercentilesFlow-3       	  300000	       527 ns/op	      96 B/op	       2 allocs/op
BenchmarkConsiderValueShortBuf-3    	  500000	       344 ns/op	       0 B/op	       0 allocs/op
BenchmarkDoSliceShortBuf-3          	  100000	      1444 ns/op	       3 B/op	       0 allocs/op
BenchmarkGetPercentilesShortBuf-3   	  200000	       564 ns/op	     136 B/op	       7 allocs/op
BenchmarkNewGaugeFloat64-3          	   20000	      6669 ns/op	    1312 B/op	      17 allocs/op
BenchmarkList-3                     	      10	  14687711 ns/op	  989904 B/op	      24 allocs/op
BenchmarkGenerateStorageKey-3       	 3000000	        52.8 ns/op	       0 B/op	       0 allocs/op
BenchmarkGet-3                      	 1000000	       121 ns/op	       0 B/op	       0 allocs/op
BenchmarkRegistry-3                 	 1000000	       187 ns/op	       0 B/op	       0 allocs/op
BenchmarkRegistryReal-3             	  200000	       685 ns/op	       0 B/op	       0 allocs/op
BenchmarkAddToRegistryReal-3        	  200000	       765 ns/op	       0 B/op	       0 allocs/op
BenchmarkRegistryRealReal-3         	  200000	      1012 ns/op	     376 B/op	       4 allocs/op
BenchmarkTagsString-3               	  300000	       462 ns/op	       0 B/op	       0 allocs/op
BenchmarkTagsFastString-3           	  500000	       233 ns/op	       0 B/op	       0 allocs/op
BenchmarkNewTimingBuffered-3        	   10000	     32947 ns/op	   84766 B/op	      53 allocs/op
BenchmarkNewTimingFlow-3            	   10000	     13458 ns/op	    3712 B/op	     103 allocs/op
```

Critical functions are `Get` (see `RegistryReal`) and `ConsiderValue`.
