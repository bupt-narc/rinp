# RINP: RINP Is Not a Proxy

RINP (RINP Is Not a Proxy) is a feasible DDoS defense solution, which can be seamlessly integrated with existing applications through its overlay-based wrap mechanism and isolated sidecar implementation. 

## Architecture

![rinp-figure](https://github.com/bupt-narc/rinp/assets/55270174/84be1216-26cd-4718-a9c6-859f60fb647a)


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


https://github.com/bupt-narc/rinp/assets/55270174/095dcab8-1150-413d-915f-4cb6b2a1ef25


Using UDP:

https://github.com/bupt-narc/rinp/assets/55270174/44913c71-c296-47da-b075-509e6c87ca47


### Result

Latency

<img width="60%" alt="lantency" src="https://github.com/bupt-narc/rinp/assets/20886330/aa21ace8-4225-4575-a9fd-37df1924fdd7"/>

Throughput

<img width="60%" alt="throughput" src="https://github.com/bupt-narc/rinp/assets/20886330/1d5b8933-35c9-4ec9-ab51-cb413a38c652"/>

Jitter

<img width="60%" alt="jitter" src="https://github.com/bupt-narc/rinp/assets/20886330/48b2a4e8-1ae9-400e-bbc1-3a756bd02a4e"/>
