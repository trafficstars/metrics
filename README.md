```
BenchmarkSortBuiltin-3                               	 5000000	       109 ns/op	      32 B/op	       1 allocs/op
BenchmarkConsiderValueFlow-3                         	 1000000	       520 ns/op	       0 B/op	       0 allocs/op
BenchmarkDoSliceFlow-3                               	  300000	      1905 ns/op	       0 B/op	       0 allocs/op
BenchmarkGetPercentilesFlow-3                        	 1000000	       540 ns/op	      96 B/op	       2 allocs/op
BenchmarkConsiderValueShortBuf-3                     	 1000000	       339 ns/op	       0 B/op	       0 allocs/op
BenchmarkDoSliceShortBuf-3                           	  300000	      1466 ns/op	       1 B/op	       0 allocs/op
BenchmarkGetPercentilesShortBuf-3                    	 1000000	       463 ns/op	     136 B/op	       7 allocs/op
BenchmarkNewGaugeFloat64-3                           	   50000	      7578 ns/op	    1248 B/op	      17 allocs/op
BenchmarkList-3                                      	      30	  15311873 ns/op	  989904 B/op	      24 allocs/op
BenchmarkGenerateStorageKey-3                        	10000000	        53.9 ns/op	       0 B/op	       0 allocs/op
BenchmarkGet-3                                       	 3000000	       119 ns/op	       0 B/op	       0 allocs/op
BenchmarkRegistry-3                                  	 2000000	       189 ns/op	       0 B/op	       0 allocs/op
BenchmarkRegistryReal-3                              	 1000000	       672 ns/op	       0 B/op	       0 allocs/op
BenchmarkAddToRegistryReal-3                         	  500000	       793 ns/op	       0 B/op	       0 allocs/op
BenchmarkRegistryRealReal_lazy-3                     	  500000	      1012 ns/op	     352 B/op	       3 allocs/op
BenchmarkRegistryRealReal_normal-3                   	  500000	       879 ns/op	      16 B/op	       1 allocs/op
BenchmarkRegistryRealReal_FastTags_withHiddenTag-3   	  500000	       934 ns/op	       0 B/op	       0 allocs/op
BenchmarkRegistryRealReal_FastTags-3                 	  500000	       891 ns/op	       0 B/op	       0 allocs/op
BenchmarkFastTag_Set-3                               	50000000	        11.1 ns/op	       0 B/op	       0 allocs/op
BenchmarkTagsString-3                                	 1000000	       463 ns/op	       0 B/op	       0 allocs/op
BenchmarkTagsFastString-3                            	 2000000	       268 ns/op	       0 B/op	       0 allocs/op
BenchmarkNewTimingBuffered-3                         	   10000	     31508 ns/op	   84699 B/op	      53 allocs/op
BenchmarkNewTimingFlow-3                             	   30000	     14079 ns/op	    3648 B/op	     103 allocs/op
```

Critical functions are `Get` (see `RegistryReal*`) and `ConsiderValue`.
