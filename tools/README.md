# __Canary Latencies Script__

## Table of Contents
1. Description
2. Prerequisites
3. Usage
4. Resources
----
## __Description__

This is a script to calculate the average producer/consumer/connection latencies from canary logs. This can be useful to determine if the histogram buckets for metrics are set optimally.

It also sorts the information into quantiles and returns them based on the number of cuts specified.

----
## __Prerequisites__

You must have access to the logs from a canary that's connected to a Kafka cluster (and configured to have the "VERBOSITY_LOG_LEVEL" turned up to "1"), either stored in files or available to be streamed in.

----
## __Usage__

The basic usage of the script: 

`python canary_latencies.py -f <your-file(s)>`

If you would like to use it with input from STDIN, remove the `-f` flag: 

`cat <your-file> | python canary_latencies.py`

### _Changing Defaults_
+ #### _Number of Cuts_
By default, the number of cuts in the [quantile method](https://docs.python.org/3/library/statistics.html#statistics.quantiles) is set to `4`. To change this, run the script with the following flag:

`python canary_latencies.py -c <your-number-of-cuts>`

+ #### _Quantile Method_
The quantile method being used is `exclusive` by default. To change this, run the script with the following flag:

`python canary_latencies.py -m inclusive`

----
## __Resources__

+ [_Python Statistics Library_](https://docs.python.org/3/library/statistics.html)
