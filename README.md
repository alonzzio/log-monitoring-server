
<div id="top"></div>


<!-- PROJECT SHIELDS -->



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

<p align="right">(<a href="#top">back to top</a>)</p>


<p align="right">(<a href="#top">back to top</a>)</p>



<!-- LICENSE -->
## License

Distributed under the MIT License. See `LICENSE.txt` for more information.

<p align="right">(<a href="#top">back to top</a>)</p>



<!-- CONTACT -->
## Contact


Project Link: [https://github.com/alonzzio/log-monitoring-server](https://github.com/alonzzio/log-monitoring-server)

<p align="right">(<a href="#top">back to top</a>)</p>
