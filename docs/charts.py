#!/usr/bin/env python3

import numpy as np
import matplotlib.pyplot as plt

# Results from "BIGCHART=1 go test -bench . -timeout 60m"
#
# BenchmarkInts/progsort/100-16     	   59418	     19674 ns/op
# BenchmarkInts/progsort/500-16     	   29150	     40506 ns/op
# BenchmarkInts/progsort/1000-16    	   18669	     63940 ns/op
# BenchmarkInts/progsort/5000-16    	    5776	    197322 ns/op
# BenchmarkInts/progsort/10000-16   	    3722	    298280 ns/op
# BenchmarkInts/progsort/50000-16   	    1065	   1098492 ns/op
# BenchmarkInts/progsort/100000-16  	     464	   2510135 ns/op
# BenchmarkInts/progsort/500000-16  	      85	  14014058 ns/op
# BenchmarkInts/progsort/1000000-16 	      51	  22202915 ns/op
# BenchmarkInts/progsort/5000000-16 	       7	 151768884 ns/op
# BenchmarkInts/progsort/10000000-16         	       4	 304130626 ns/op
# BenchmarkInts/progsort/50000000-16         	       1	2277216824 ns/op
# BenchmarkInts/progsort/100000000-16        	       1	3981873534 ns/op
# BenchmarkInts/progsort/500000000-16        	       1	19480695089 ns/op
# BenchmarkInts/progsort/1000000000-16       	       1	42395064737 ns/op
# BenchmarkInts/stdlib/100-16                	  709266	      1453 ns/op
# BenchmarkInts/stdlib/500-16                	  114210	     10621 ns/op
# BenchmarkInts/stdlib/1000-16               	   49429	     23992 ns/op
# BenchmarkInts/stdlib/5000-16               	    7960	    143065 ns/op
# BenchmarkInts/stdlib/10000-16              	    3766	    307078 ns/op
# BenchmarkInts/stdlib/50000-16              	     674	   1771503 ns/op
# BenchmarkInts/stdlib/100000-16             	     313	   3764269 ns/op
# BenchmarkInts/stdlib/500000-16             	      50	  21737094 ns/op
# BenchmarkInts/stdlib/1000000-16            	      25	  47554151 ns/op
# BenchmarkInts/stdlib/5000000-16            	       2	 502221224 ns/op
# BenchmarkInts/stdlib/10000000-16           	       1	1614043333 ns/op
# BenchmarkInts/stdlib/50000000-16           	       1	8444168939 ns/op
# BenchmarkInts/stdlib/100000000-16          	       1	18037104155 ns/op
# BenchmarkInts/stdlib/500000000-16          	       1	93730488701 ns/op
# BenchmarkInts/stdlib/1000000000-16         	       1	216609456504 ns/op

# Values in nanoseconds
PROG = [19674,40506,63940,197322,298280,1098492,2510135,14014058,22202915,151768884,304130626,2277216824,3981873534,19480695089,42395064737]
STD = [1453,10621,23992,143065,307078,1771503,3764269,21737094,47554151,502221224,1614043333,8444168939,18037104155,93730488701,216609456504]
LABELS = ['100','500','1k','5k','10k','50k','100k','500k','1m','5m','10m','50m','100m','500m','1b']

# Convert nanoseconds to seconds
PROG = np.divide(PROG,1e9)
STD = np.divide(STD,1e9)

def chart(start, end, name):  
	# Take a slice of the 
	PROGP = PROG[start:end]
	STDP = STD[start:end]
	LABELSP = LABELS[start:end]

	# Set width of bar
	barWidth = 0.25
	fig = plt.subplots(figsize =(12, 8))

	# Set position of bar on X axis
	br1 = np.arange(len(PROGP))
	br2 = [x + barWidth for x in br1]
	br3 = [x + barWidth for x in br2]

	plt.bar(br1, PROGP, color ='r', width = barWidth, edgecolor ='grey', label ='progsort.Sort')
	plt.bar(br2, STDP, color ='g', width = barWidth, edgecolor ='grey', label ='sort.Ints')
	plt.xlabel('Number of Ints', fontweight ='bold', fontsize = 15)
	plt.ylabel('Seconds', fontweight ='bold', fontsize = 15)
	plt.xticks([r + barWidth for r in range(len(PROGP))], LABELSP)

	plt.legend()
	plt.savefig(name)
	# plt.show()

# Output each chart
chart(0, 5, "chart1.png")
chart(5, 10, "chart2.png")
chart(10, 15, "chart3.png")
