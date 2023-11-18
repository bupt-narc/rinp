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

## Evaluation

### Environment

Evaluation is conducted between Bejing, China and Guangzhou, China with a fixed bandwidth of 10Mbps.

We deploy an `Authenticator`, 2 proxies, a scheduler, and a database on host machines with 2 cores of AMD EPYC 7K62 and 4 GB RAM in Guangzhou. We deploy some `Accessors` in Linux virtual machines in Beijing, each of which has 2 CPUs from a Xeon Silver 4216 CPU. As is known, Beijing and Guangzhou are more than 2100 kilometers apart, which can represent the real network status.

Using TCP:

https://github.com/bupt-narc/rinp/assets/55270174/898364df-e3bf-460a-813c-3748034aecdc

Using UDP:

https://github.com/bupt-narc/rinp/assets/55270174/d6be6529-9f17-4ead-8422-dcce909aa0f1

### Result

Latency

<img width="60%" alt="lantency" src="https://github.com/bupt-narc/rinp/assets/20886330/7fa2d57c-89b4-4ae7-82c9-888c18f6c8da"/>

Throughput

<img width="60%" alt="throughput" src="https://github.com/bupt-narc/rinp/assets/20886330/a96ef6fd-3f12-4861-a272-d016916e5f2d"/>

Jitter

<img width="60%" alt="jitter" src="https://github.com/bupt-narc/rinp/assets/20886330/0be46ab2-848f-462d-b91b-06e09b6a77a7"/>
