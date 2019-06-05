# Herdius

***Herdius*** is a P2P network and Byzantine Fault Tolerant blockchain. The Herdius blockchain is specifically tailored for fast transaction settlement. It is the backbone of all products built at Herdius including wallet, trading enginer, interoperability, etc. Our aim is to create an inter-blockchain gateway and so called "highway" through which the whole of decentralised web becomes accessible. 

## Current architecture

One of the highlights of the Herdius blockchain is the **State-split and transaction batching mechanism** we call these two processes **BoB** Blocks-on Blocks. Currently a single node called the **Supervisor** receives all transactions from clients, places them in different batches, then assigns them to different **Validator Groups** creating so called **Child-blocks** in the process. These **Child-blocks** are linked to a **Base-block** and form a trie together, where each individual **Child-block** corresponds to their own State and network shard. Once **Validator groups** receive their assigned **Child-block** each of the validators within the group proceeds to validate each individual transaction inside their batch. Once validators reach consesus among their peers in their own validator groups, they transmit the result of the agreement (consensus) back to the **Supervisor**. Once all the responses are collected, the **Supervisor** finalizes the block with all of its **Child-blocks** which in turn becomes a block in the blockchain.

The key innovation of our approach is the fact that each individual Child-block or shard is **only** being validated by the validator group that it was assigned to. Validation of each individaul child block, each assigned to different validator group, runs in parallel. As the validator set (which is at all times maintained by the Supervisor) grows, so does the speed and the amount of transactions the Herdius chain can handle. This is a result of the supervisor dividing transactions into optimal batches for the validator groups. Thanks to us maintaining a single State with the Supervisor, cross-shard communication and transactions are not an issue. Every address can make transactions seamlessly with each other. The system is written in [Go](https://golang.org/) by Engineers at **Herdius**. For protocol architecture details, see [Herdius Technical Paper](https://herdius.com/whitepaper/Herdius_Technical_Paper.pdf) as well as our [Medium Blog](https://medium.com/herdius). 
 
## Terms


## Minimum requirements

Requirements|Detail
---|---
Go version | [Go 1.11.4 or higher](https://golang.org/dl/)
Proctc | [Protobuf compiler](https://github.com/google/protobuf/releases)
Make | [make tool](http://www.gnu.org/software/make/)

## Quick Start the Service Locally

#### Get Source Code

```
mkdir -p $GOPATH/src/github.com/herdius
git clone https://github.com/herdius/herdius-core.git $GOPATH/src/github.com/herdius/
cd $GOPATH/src/github.com/herdius/herdius-core
```

#### Create State Trie and Blockchain DB Dirs

```
make create_db_dirs
```

#### Install and Build

```
export GO111MODULE=on
make install
```

#### Test

```
make run-test
```

#### Start Supervisor Server

```
make start-supervisor
```

Supervisor server will start at **tcp://127.0.0.1:3000**

#### Start Validator Server

```
make start-validator PORT=3001 PEERS="tcp://127.0.0.1:3000" HOST="127.0.0.1"
```

Validator node will be bootstrapped with Supervisor node and it will be hosted at **tcp://127.0.0.1:3001**. And in the same way multiple validator nodes could be bootstrapped at various ports and hosts.

#### Sample output

```
2:57PM INF Last Block Hash : 812B7A12C8E774C6EE5D5B3F76623102EED61A2956B632BCABA7A5E367EBBAB9
2:57PM INF Height : 0
2:57PM INF Timestamp : 2019-02-08 14:57:30 +0530 IST
2:57PM INF State root : 4A03B15DFAE8C35E5A52C170AAF8749A10C14178A6BD21F3D5F43D7C80D91476
2:57PM INF <tcp://127.0.0.1:3001> Connection established

2:57PM INF New Block Added
2:57PM INF Block Id: 1DCB99A379E851C065AE99F77A4689E5B6A066CBDAD359C9E375660052C6A151
2:57PM INF Block Height: 1
2:57PM INF Timestamp : 2019-02-08 14:57:30 +0530 IST
2:57PM INF State root : E6C27D9BDE675F2212937FD36CF3917CAE038765262EBB196B1571FBB2CC8EBC
2:57PM INF Total time : 148.26511ms
```

## Guidelines on AWS EC2 setup

#### Create an AWS EC2 Supervisor

We have created and configured Herdius testnet and performed testing on **t3.small** EC2 instances with each instance having 2GB RAM and 2 CPU Cores. The following instructions and examples were performed on a Centos x84-64 machine. However, Herdius setup could be configured on any  EC2 instance.

#### Install Tools and Dependecies

```
sudo yum update
sudo yum install -y git gcc jq
```

[Install Go](https://golang.org/doc/install)


#### Clone from git repository

```
mkdir -p $GOPATH/src/github.com/herdius/
git clone https://github.com/herdius/herdius-core.git $GOPATH/src/github.com/herdius/
cd $GOPATH/src/github.com/herdius/herdius-core/
```

#### Herdius Core installation

```
make all
```

#### (Optional) Create a Validator

The validator process requires the same setup as the Supervisor, above. In order to prepare a Validor(s) to join the cluster with the Supervisor, please follow the same above steps.

#### Secret Store and Retrieval

The Supervisor requires access to external blockchains, and thus needs API keys and tokens. At the moment, these secrets are read from environment variables. Export these values to the environment, for the Supervisor to use:

```
export INFURAID=<infura id here>
export BLOCKCHAIN_INFO_KEY=<info key here>
```

#### Start Supervisor Server

```
make start-supervisor ENV=staging

Starting supervisor node -groupsize=3 -waitTime=3 -env=staging
2019/06/03 10:54:49 value.go:758: Replaying from value pointer: {Fid:0 Len:0 Offset:0}
2019/06/03 10:54:49 value.go:771: Iterating file id: 0
2019/06/03 10:54:49 value.go:774: Iteration took: 32.644µs
2019/06/03 10:54:49 utils.go:82: No peers discovered in network, retrying
10:54AM INF Listening for peers. address=tcp://127.0.0.1:3000
2019/06/03 10:54:49 value.go:758: Replaying from value pointer: {Fid:0 Len:0 Offset:0}
2019/06/03 10:54:49 value.go:771: Iterating file id: 0
2019/06/03 10:54:49 value.go:774: Iteration took: 12.671µs
10:54AM INF Last Block Hash: 812B7A12C8E774C6EE5D5B3F76623102EED61A2956B632BCABA7A5E367EBBAB9
10:54AM INF Height: 0
10:54AM INF Timestamp: 2019-06-03 10:54:49 +0200 CEST
10:54AM INF State root: CB5D298E8E25DBC8210889D8D31AC7A956384444024EC08A2FEF704251094542
2019/06/03 10:54:49 account.go:33: Infura Url with Project ID: https://ropsten.infura.io/v3/
2019/06/03 10:54:52 utils.go:82: No peers discovered in network, retrying
2019/06/03 10:54:52 service.go:346: Block creation wait time (3) elapsed, creating singular base block but with 0 transactions
2019/06/03 10:54:52 service.go:429: Total Accounts to update 0
2019/06/03 10:54:52 service.go:355: nonfatal: failed to backup new block to S3: failure searching S3 for previous block backup: could not list previous block in S3: MissingRegion: could not find region configuration
10:54AM INF New Block Added
10:54AM INF Block Id: 327D8DA50B49851EC8794506F31DA6ABBDE1EF2A0F0260F0329EF223CD0D9D61
10:54AM INF Last Block Id: 812B7A12C8E774C6EE5D5B3F76623102EED61A2956B632BCABA7A5E367EBBAB9
10:54AM INF Block Height: 1
10:54AM INF Timestamp : 2019-06-03 10:54:49 +0200 CEST
10:54AM INF State root : CB5D298E8E25DBC8210889D8D31AC7A956384444024EC08A2FEF704251094542

```

#### (Optional) Start Validator Node

From the Validator node:

```
cd $GOPATH/src/github.com/herdius/herdius-core/
make start-validator HOST="<HOST-IP>" PEERS="tcp://<SUPERVISOR-IP>:3000" ENV=staging

Starting validator node
12:33PM INF Listening for peers. address=tcp://<HOST-IP>:3000
```

## Backup and Resiliency

The Herdius blockchain is designed such that if at any given time, a single Supervisor is connected to the network and addressable, then the chain persists. However, in the extremely unlikely scenario where all Supervisor's go offline, a number of processes will gracefully fail.
While this can problematic, it is important that the chain must persist on past this event. Thus, the entirety of the staging and production block chain is persisted to S3 upon every new block creation. This ensures a high degree of resiliency, as S3 guarantees 99.999999999% durability of objects.

## Contributing

Thank you for your interest in advancing the development of the Herdius Blockchain! :heart: :heart: :heart:

The Herdius Blockchain is an open-source project that welcomes source-code development, discovered issues, and recommendations on our blockchain and product's future! The following are some required and suggested contribution instructions:

#### Table of Contents

* [Major technologies](#major-technologies)
* [Repository structure](#repository-structure)
* [Documentation](#documentation)
* [Pull Request Instructions](#pull-request-instructions)

### Major Technologies

### Repository structure

### Documentation

### Pull Request Instructions

All new pull requests should be submitted to `master` in GitHub. Stable releases are marked with a git tag and are indicated with a preceding `v` and then the release number (eg. `v1.0.0`).

For a pull request to be review, apply the GitHub "Ready for Review" label.

All pull requests must pass the CI/CD pipeline testing prior to merge. This pipeline is integrated with GitHub and will report success or failure as a GitHub status check.


## License

The herdius-core library (i.e. all files and code without existing reference to other, separate license) is licensend under the [GNU Affero General Public License v3.0](https://www.gnu.org/licenses/licenses.html#AGPL). Text of the license file is also include in the root folder under LICENSE. 


