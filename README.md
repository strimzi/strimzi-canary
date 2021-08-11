[![Acceptance](https://github.com/strimzi/strimzi-canary/actions/workflows/acceptance.yml/badge.svg)](https://github.com/strimzi/strimzi-canary/actions/workflows/acceptance.yml)
[![License](https://img.shields.io/badge/license-Apache--2.0-blue.svg)](http://www.apache.org/licenses/LICENSE-2.0)
[![Twitter Follow](https://img.shields.io/twitter/follow/strimziio.svg?style=social&label=Follow&style=for-the-badge)](https://twitter.com/strimziio)

# Strimzi canary

This repository contains the Strimzi canary tool implementation.
It acts as an indicator of whether Kafka clusters are operating correctly.
This is achieved by creating a canary topic and periodically producing and consuming events on the topic and getting metrics out of these exchanges.
