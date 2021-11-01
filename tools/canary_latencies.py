#!/usr/bin/env python3

#
# Copyright Strimzi authors.
# License: Apache License 2.0 (see the file LICENSE or http://apache.org/licenses/LICENSE-2.0.html).
#

import re
import argparse
import fileinput
import statistics

def parse_logs(files):
    producer_latencies = []
    e2e_latencies = []
    connection_latencies = []

    producer_pattern = re.compile(".*?producer.go:\d+\]\sMessage\ssent.*duration=([0-9]+)\sms")
    e2e_pattern = re.compile(".*?consumer.go:\d+\]\sMessage\sreceived.*duration=([0-9]+)\sms")
    connection_pattern = re.compile(".*?connection_check.go:\d+\].*broker\s[0-9]\sin\s([0-9]+)\sms")

    for line in fileinput.input(files):
            if match := producer_pattern.match(line):
                producer_latencies.append(int(match.group(1)))
            elif match := e2e_pattern.match(line):
                e2e_latencies.append(int(match.group(1)))
            elif match := connection_pattern.match(line):
                connection_latencies.append(int(match.group(1)))

    return producer_latencies, e2e_latencies, connection_latencies

def calculate_quantiles(latencies, quantileMethod, numberOfCuts):
    quantiles = statistics.quantiles(latencies, n=numberOfCuts, method=quantileMethod)
    return [round(p, 1) for p in quantiles]

if __name__ == '__main__':
    parser = argparse.ArgumentParser()
    parser.add_argument('-f', '--files', nargs='+', default=[], help='log files (if empty, expects logs to be piped through stdin)')
    parser.add_argument('-m', '--method', default='exclusive', choices=['exclusive', 'inclusive'], metavar='exclusive', help='change quantile method from exclusive to inclusive')
    parser.add_argument('-c', '--cuts', default=4, type=int, metavar='4', help='number of cuts in quantile method')
    args = parser.parse_args()

    producer_latencies, e2e_latencies, connection_latencies = parse_logs(args.files)

    print("\nProducer Latency Average: ")
    print(round(statistics.mean(producer_latencies), 1))

    print("\nE2E Latency Average: ")
    print(round(statistics.mean(e2e_latencies), 1))

    print("\nConnection Latency Average: ")
    print(round(statistics.mean(connection_latencies),1 ))

    quantile_method = args.method
    number_of_cuts = args.cuts

    print("\nProducer Quantiles: ")
    print(calculate_quantiles(producer_latencies, quantile_method, number_of_cuts))

    print("\nE2E Quantiles: ")
    print(calculate_quantiles(e2e_latencies, quantile_method, number_of_cuts))

    print("\nConnection Quantiles: ")
    print(calculate_quantiles(connection_latencies, quantile_method, number_of_cuts))
