# Herdius

***Herdius*** is a P2P network and Byzantine Fault Tolerance blockchain system. It has the capability of vertically stretching itself, thereby fitting and stacking several blocks on top of each other in the process. This stretching mechanism allows us to introduce parallelization to the validation process and is our solution to make transactions within the Herdius network fast and scalable. This stretching mechanism is called Blocks-of-Blocks (**BoB**). Herdius also divides the P2P network into shards or better termed as validator groups where in each group is capable of validating the transactions in parallel. This system is written in [Go](https://golang.org/) by Developers at **Herdius**.

For protocol architecture details, see [Herdius Technical Paper](https://herdius.com/whitepaper/Herdius_Technical_Paper.pdf).

## Minimum requirements

Requirements|Detail
---|---
Go version | [Go1.11.4 or higher](https://golang.org/dl/)
Proctc | [Protobuf compiler](https://github.com/google/protobuf/releases)
Make | [make tool](http://www.gnu.org/software/make/)

## Quick Start the Service Locally

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
```

### Test

```
make run-test
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

**NOTE**: Make sure all peer connections are established before presssing **"y"**.
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

## Guidelines on AWS EC2 setup

### Create an AWS EC2 Instance

We have created and configured Herdius testnet and performed testing on **t3.small** EC2 instances with each instance having 2GB RAM and 2 CPU Cores. However, Herdius setup could be configured on any free tier EC2 instances.

### Install Tools and Dependecies
```
sudo yum update
sudo yum install git
sudo yum install gcc
mkdir downloads && cd downloads
sudo wget https://dl.google.com/go/go1.11.5.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.11.5.linux-amd64.tar.gz
```

### Create required directories and set permissions

```
mkdir $HOME/go_projects
sudo chmod 777 -R $HOME/go_projects/
```

### Setup bashrc
```
sudo nano ~/.bashrc
```

Enter the below details and save the file

```
export PATH=$PATH:/usr/local/go/bin
export GOPATH="$HOME/go_projects"
export PATH=$PATH:$GOROOT/bin:$GOPATH/bin
```

Execute the below command
```
source ~/.bashrc
```

Check go version
```
go version
go version go1.11.5 linux/amd64
```

### Clone from git repository

Exectue the below command
```
cd /usr/local/go/src && sudo mkdir github.com && cd github.com && sudo mkdir herdius && cd herdius
```

Clone the herdius core
```
sudo git clone https://github.com/herdius/herdius-core.git
```

### Give permission to required folders

```
sudo chmod 777 -R /home/ec2-user/go_projects && sudo chmod 777 -R herdius-core/supervisor/testdata/ && sudo chmod 777 -R /usr/local/go/pkg
sudo chmod 777 -R /usr/local/go/src/github.com/herdius/herdius-core/
```

### Herdius Core installation

```
cd herdius-core/
make all
```



Exactly the same above setup guidelines need to be followed for other EC2 instances or nodes where core will be executing as validator. And once all the setups are completed. Run the below commands.



### Start Supervisor Server

```
cd /usr/local/go/src/github.com/herdius/herdius-core/
make start-supervisor
```

Once the supervisor node is started, it will create the genesis block and it will start listening at port **3000**.

```
Starting supervisor node
11:41AM INF Listening for peers. address=tcp://<HOST-IP>:3000
2019/02/19 11:41:13 Replaying from value pointer: {Fid:0 Len:0 Offset:0}
2019/02/19 11:41:13 Iterating file id: 0
2019/02/19 11:41:13 Iteration took: 15.354µs
11:41AM INF Last Block Hash : 812B7A12C8E774C6EE5D5B3F76623102EED61A2956B632BCABA7A5E367EBBAB9
11:41AM INF Height : 0
11:41AM INF Timestamp : 2019-02-19 11:41:13 +0000 UTC
11:41AM INF State root : 4A03B15DFAE8C35E5A52C170AAF8749A10C14178A6BD21F3D5F43D7C80D91476
11:41AM INF Please press 'y' to load transactions from file. 

```

### Start Validator Node

Open another EC2 instance to listen to a validator node

```
cd /usr/local/go/src/github.com/herdius/herdius-core/
make start-validator HOST="<HOST-IP>" PEERS="tcp://<SUPERVISOR-IP>:3000"
```

Once the validator node is started, it will start listening at port **3000**.

```
Starting validator node
12:33PM INF Listening for peers. address=tcp://<HOST-IP>:3000

```



