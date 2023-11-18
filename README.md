# RINP: RINP Is Not a Proxy

RINP (RINP Is Not a Proxy) is a feasible DDoS defense solution, which can be seamlessly integrated with existing applications through its overlay-based wrap mechanism and isolated sidecar implementation. 

## Architecture

![rinp-figure](https://github.com/bupt-narc/rinp/assets/55270174/15d3e314-149d-4d79-8fd9-c83db5edd8e3)

## Quick Start

Make sure you Docker, and GNU-Make installed and running on a Linux machine.

1. Build a base container image which is useful for testing purposes: `cd examples && make && cd -`
2. Build RINP components using the base container that we just built: `BASE_IMAGE=netutils make container`
3. Start RINP: `cd examples && docker compose up`. Check for any errors.
4. (In a separate terminal) Run a iperf server to test with: `docker run -it service iperf3 -s`
5. (In a separate terminal) Run a iperf client to test: `docker run -it user iperf3 -c 11.22.33.44`

Notice the client used an IP that is virtual (meaning RINP is functioning). If nothing goes wrong, you should see the test going. Feel free to raise an issue if you have questions.

## Evaluation Videos

Evaluation is conducted between Bejing, China and Guangzhou, China with a fixed bandwidth of 10Mbps.

Using TCP:

https://github.com/bupt-narc/rinp/assets/55270174/898364df-e3bf-460a-813c-3748034aecdc

Using UDP:

https://github.com/bupt-narc/rinp/assets/55270174/d6be6529-9f17-4ead-8422-dcce909aa0f1
