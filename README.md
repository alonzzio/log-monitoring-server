
<div id="top"></div>


<br />
<div align="center">

<h2 align="center">LOG MONITORING  SERVER (LMS)</h2>

  <p align="center">
    Fiskil Challenge Project!
    <br />

</div>

[![CircleCI](https://circleci.com/gh/alonzzio/log-monitoring-server/tree/master.svg?style=svg)](https://circleci.com/gh/alonzzio/log-monitoring-server/tree/master)

<!-- ABOUT THE PROJECT ( LMS )-->
## About The Project (LMS)

Log Monitoring Server fetches Logs from Services like Pub/Sub.Stream pull the message,make it as batch and save to Database

<p align="right">(<a href="#top">back to top</a>)</p>


<!-- GETTING STARTED -->

## Project structure:
There are three separate modules in the frameworks.
* PUB/SUB Server
* Data Collection Layer
* Data Access Layer

```
log-monitoring-server
├── .circleci
├── cmd
│   ├── env
│   │   ├── dal.env
│   │   ├── database.env
│   │   ├── dcl.env
│   │   └── pubsub.env
│   ├── Makefile    
│   ├── main.go
│   ├── routes.go
│   └── run.go
├── internal
│    ├── accesess
│    │   ├── access.go
│    │   └── api.go
│    ├── collection
│    │   ├── collection.go
│    │   ├── db.go
│    │   └── pools.go
│    ├── config
│    │   ├── config.go
│    │   └── environment.go
│    ├── db
│    │   └── db.go
│    ├── helpers
│    │   └── db.go
│    └── pst
│        ├── process.go
│        ├── pst.go
│        ├── pst_test.go
│        └── server.go
├── LMSlogs
│    ├── app.log
│    └── system.log
├── .gitignore
├── docker-compose.yml
├── go.mod
│    └── go.sum
├── LICENSE
└── README.md
```
### Runtime:

Frame works using `Repository pattern` for Global variable sharing.
All `internal` packages are inside `internal/`
Working directory under `cmd/`
Env files directory `cmd/env/`

Every internal package has embedded Repository struct and all functions are method of `Repository`

When app init repository will initialise and share its `pointer` across packages.

Processing `pub/sub` is a independent service which will continuously publish messages to pub/sub.(its `frequency` and `message per publish` and number of publish `workers` pool can be configured through `env`)

Processing `Data Collection Layer` reads messages from `pub/sub` and send to a buffered `channel`. Another process will make queues and sort to message and severity together send to another channel.
A BD process will read this processed messages and severity from the channel and insert to `DB` as a `transaction`.If any fail from DB. It will retry. and a statics log will write to `app.log` file. `(LMSlogging/)`
All other System logs will write to `system.log`.

`Message per batch` number of concurrent `workers` can ve configured through env.

* _docker test is not used in this project. However, MYSQL docker image is used._
* _Docker-compose.yml will take care of Mysql and create lms database._
* _Every re-run of the project will truncate the data from table for testing_

### Prerequisites

There should be Golang and Docker installed
* Go
* Docker
* Git


### Git Clone Project

_Clone the project from github._

1. Find the project at [alonzzio/log-monitoring-server](https://github.com/alonzzio/log-monitoring-server)
2. Clone the repo
   ```sh
   git clone https://github.com/alonzzio/log-monitoring-server.git
   ```
3. Navigate to the working directory
   ```sh
   cd cmd
   ```
   1. Test,Build and Run using `Makefile`

      1. Docker Compose download and run Mysql image
         ```sh
         make docker-up
         ```
      2. Download Dependencies for Golang
         ```sh
         make dep
         ```
      3. Test
         ```sh
         make test
         ```
      4. Build
         ```sh
         make build
         ```
      5. Run
         ```sh
         make run
         ``` 
      6. API testing
         Where SN argument is `Service-Name:` and `S` is severity
         ```sh
         make run-api SN="1" S="Info"
         ``` 
         We can try different numbers and Severity to in the arguments.
      
         Result:
         ```shell
            {"status":200,"status_text":"Service and Severity count match! OK","services":{"severity_name":"Service-name:1","service_severity":"Info","count":2},"services_severity":{"severity_name":"Service-name:1","service_severity":"Info","count":2}}
          ```

      7. Do everything together (No API Test)
         ```sh
         make all
         ``` 
      8. Docker kill container
         ```sh
         make docker-kill
         ``` 
      9. Clean docker container
         ```sh
         make docker-clean
         ``` 

<p align="right">(<a href="#top">back to top</a>)</p>


## Configure Project ENV

WE can configure most of the Environment variables `./cmd/env/*.env` files

* Number of workers
* Message Batch 
* Message Payload Configurations (message length paragraph count etc.)
* Pub/Sub Configurations
* Other general Configurations.

_Note:Multiple env files is supported. Extensions of the file should be `*.ENV`_

<p align="right">(<a href="#top">back to top</a>)</p>

## Data Access Layer
Server running at local host port `8080` address: `localhost:8080` Port number can be configured though env files.

ping is just a ping to the server
eg:
`cURL`

   ```sh
   curl -X GET \
   http://localhost:8080/ping \
   -H 'cache-control: no-cache' \
   ``` 
Response Will Be:

   ```json
{
    "status": 200,
    "status_text": "Welcome to Data Access Layer"
}
   ```

To get all Service names in the DB:
`"severity_name"` is `"Service-name:2"` and
`Severity` as `Info`

   ```shell
curl --location --request GET 'localhost:8080/service-severity-stat?service-name=Service-name:2&severity=Info' \
--data-raw ''
   ```

Sample Response if found:

```json
{
   "status": 200,
   "status_text": "Service and Severity count match! OK",
   "services": {
      "severity_name": "Service-name:2",
      "service_severity": "Info",
      "count": 8
   },
   "services_severity": {
      "severity_name": "Service-name:2",
      "service_severity": "Info",
      "count": 8
   }
}
   ```

If not found:
```json
{
    "status": 200,
    "status_text": "Service and Severity not found! OK"
}
```
If service name not supplied:
```json
{
    "status": 400,
    "status_text": "Service name not supplied"
}
```
If severity not supplied:
```json
{
   "status": 400,
   "status_text": "severity not supplied"
}
```
<p align="right">(<a href="#top">back to top</a>)</p>
