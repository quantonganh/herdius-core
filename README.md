# Herdius

***Herdius*** is a P2P network and Byzantine Fault Tolerance blockchain system. It has the capability of vertically stretching itself, thereby fitting and stacking several blocks on top of each other in the process. This stretching mechanism allows us to introduce parallelization to the validation process and is our solution to make transactions within the Herdius network fast and scalable. This stretching mechanism is called Blocks-of-Blocks (**BoB**). Herdius also divides the P2P network into shards or better termed as validator groups where in each group is capable of validating the transactions in parallel. This system is written in [Go](https://golang.org/) by Developers at **Herdius**.

For protocol architecture details, see [Herdius Technical Paper](https://herdius.com/whitepaper/Herdius_Technical_Paper.pdf).

## Minimum requirements

Requirements|Detail
---|---
Go version | [Go1.11.4 or higher](https://golang.org/dl/)
Proctc | [Protobuf compiler](https://github.com/google/protobuf/releases)
Make | [make tool](http://www.gnu.org/software/make/)

## Quick Start

### Get Source Code

```
mkdir -p $GOPATH/src/github.com/herdius
cd $GOPATH/src/github.com/herdius
git clone https://github.com/herdius/herdius-core
cd herdius-core
```

### Create State Trie and Blockchain DB Dirs

```
make create_db_dirs
```

### Install and Build

```
make install
make build
```

### Test

```
make run_test
```

### Start Supervisor Server

```
make start-supervisor
```

Supervisor server will start at **tcp://127.0.0.1:3000**

### Start Validator Server

```
make start-validator PORT=3001 PEERS="tcp://127.0.0.1:3000" HOST="127.0.0.1"
```

Validator node will be bootstrapped with Supervisor node and it will be hosted at **tcp://127.0.0.1:3001**. And in the same way multiple validator nodes could be bootstrapped at various ports and hosts.

Once the initial setup is done, at the supervisor console it will ask if the transactions are to be loaded from a file. Currently we have a [file](supervisor/testdata/txs.json) which has 3000 transactions. However, # of transactions can be uploaded of any count.

Please press **"y"** to load transacions from the file.

```
2:57PM INF Last Block Hash : 812B7A12C8E774C6EE5D5B3F76623102EED61A2956B632BCABA7A5E367EBBAB9
2:57PM INF Height : 0
2:57PM INF Timestamp : 2019-02-08 14:57:30 +0530 IST
2:57PM INF State root : 4A03B15DFAE8C35E5A52C170AAF8749A10C14178A6BD21F3D5F43D7C80D91476
2:57PM INF <tcp://127.0.0.1:3001> Connection established
2:57PM INF Please press 'y' to load transactions from file. 
y

```
Once the new base block is created, below details will appear in the console.

```
2:57PM INF New Block Added
2:57PM INF Block Id: 1DCB99A379E851C065AE99F77A4689E5B6A066CBDAD359C9E375660052C6A151
2:57PM INF Block Height: 1
2:57PM INF Timestamp : 2019-02-08 14:57:30 +0530 IST
2:57PM INF State root : E6C27D9BDE675F2212937FD36CF3917CAE038765262EBB196B1571FBB2CC8EBC
2:57PM INF Total time : 148.26511ms
```


