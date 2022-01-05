
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
├── logs
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

The framework has package `main`  which is under `cmd/`
All other packages in `internal` folder. All application wide configuration is in `internal/config` This package utilises repository pattern for other packages and `main`

Program entry point is in `cmd/main.go`
package main make sure connecting to `database` and loading all `env` variables to `AppConfig` which holds all application wide configurations.
`ENV` files can be found under `cmd/env/*.env` we can add more files if we want.

Once all `env` variables is loaded, then `main` will initialise `pubsub Fake server` which is package `pst` `(pub/sus test)`. It will run in a separate goroutine and runs it forever listening for `pub/sub` 

After that `main` will call a `InitPubSubProcess` `func` from `pst` package, Which initialise `service-name pools` of given size reads from `env` variables.
A number of `goroutines` will start `publishing` messages to pub/sub. This service will run as separate process.

At this same time `Data Collection Layer`  initialise a `buffered` channels  `ReceiverJobs` , `ReceiverResult` , `LogsResult`.
* A process will continuously send jobs to `ReceiverJobs` channel
* `ReceiverResult` channel will receive messages from pub/sub. Process as batch messages and send to `LogsResult` channel Which includes `*[]Messages` and `*[]ServiceSeverity`
* `MessageDbProcessWorker` will receive from `LogsResult` channel and `batch` insert to database as `transaction`. There will be 5 `retries` if error occurred.If all retries fail Batch message will send back to `LogsResult` channel 

There is another `concurrent` web server will start at port `8080` for `Data Access Layer`


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
4. Test,Build and Run using `Makefile`

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
   6. Do everything together
      ```sh
      make all
      ``` 
   7. Docker kill container
      ```sh
      make docker-kill
      ``` 
   8. Clean docker container
      ```sh
      make docker-clean
      ``` 

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

   ```sh
   Welcome to Data Access Layer% 
   ```

   To get all Service names in the DB:

   ```shell
   curl -X GET \
     http://localhost:8080/services \
     -H 'cache-control: no-cache' \
   ```

   Response:

```json
   {
       "status": 200,
       "status_text": "ok",
       "services": [
           {
               "name": "Service-name:1"
           },
           {
               "name": "Service-name:10"
           },
           {
               "name": "Service-name:2"
           },
           {
               "name": "Service-name:3"
           },
           {
               "name": "Service-name:4"
           },
           {
               "name": "Service-name:5"
           },
           {
               "name": "Service-name:6"
           },
           {
               "name": "Service-name:7"
           },
           {
               "name": "Service-name:8"
           },
           {
               "name": "Service-name:9"
           }
       ]
   }
   ```
 _Note: This is just a demo service names_

### Data Access Layer API is being written.
Comparing or Analytical API's is not Ready




<!-- Configure -->
## Configure Project ENV

WE can configure most of the Environment variables `./cmd/env/*.env` files

* Number of workers
* Message Batch 
* Message Payload Configurations (message length paragraph count etc.)
* Pub/Sub Configurations
* Other general Configurations.

_Note:Multiple env files is supported. Extensions of the file should be `*.ENV`_

<p align="right">(<a href="#top">back to top</a>)</p>


## Testing Files not complete.
I Will add more testing in coming days


<!-- Repository -->
## Repository


Project Link: [https://github.com/alonzzio/log-monitoring-server](https://github.com/alonzzio/log-monitoring-server)

<p align="right">(<a href="#top">back to top</a>)</p>
