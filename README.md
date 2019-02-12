goos: linux
goarch: amd64
pkg: github.com/trafficstars/fastmetrics
BenchmarkSortBuiltin-3                  	 1000000	      1694 ns/op	      33 B/op	       1 allocs/op
BenchmarkConsiderValueFlow-3            	 2000000	       649 ns/op	       0 B/op	       0 allocs/op
BenchmarkDoSliceFlow-3                  	  500000	      3411 ns/op	       0 B/op	       0 allocs/op
BenchmarkConsiderValueShortBuf-3        	 3000000	       376 ns/op	       0 B/op	       0 allocs/op
BenchmarkDoSliceShortBuf-3              	 1000000	      1562 ns/op	       3 B/op	       0 allocs/op
BenchmarkGetPercentilesShortBuf-3       	    3000	    410182 ns/op	     191 B/op	       8 allocs/op
BenchmarkList-3                         	     100	  16323449 ns/op	  989904 B/op	      24 allocs/op
BenchmarkGenerateStorageKey-3           	 5000000	       274 ns/op	       0 B/op	       0 allocs/op
BenchmarkGet-3                          	 5000000	       288 ns/op	       0 B/op	       0 allocs/op
BenchmarkRegistry-3                     	 5000000	       439 ns/op	       0 B/op	       0 allocs/op
BenchmarkRegistryReal-3                 	 1000000	      1282 ns/op	       0 B/op	       0 allocs/op
BenchmarkAddToRegistryReal-3            	 1000000	      1438 ns/op	       0 B/op	       0 allocs/op
BenchmarkRegistryReal_withHiddenTag-3   	 2000000	       920 ns/op	       0 B/op	       0 allocs/op
BenchmarkRegistryReal_FastTags-3        	 2000000	       881 ns/op	       0 B/op	       0 allocs/op
BenchmarkTagsString-3                   	 1000000	      1117 ns/op	       0 B/op	       0 allocs/op
BenchmarkTagsFastString-3               	 2000000	       585 ns/op	       0 B/op	       0 allocs/op
BenchmarkTimingFillStats-3              	    2000	   1237595 ns/op	 1768659 B/op	      68 allocs/op
PASS
ok  	github.com/trafficstars/fastmetrics	33.785s
