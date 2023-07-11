# BBQueue
A simple and efficient queueing service implemented in Golang with Redis as the storage backend. BBQueue is designed to handle more than 35,000 requests per second and more than 20,000 concurrent request with no errors, and costs don't scale with usage.

## Prerequisites
Before you begin, ensure you have met the following requirements:

* You have a working Docker installation. If not, you can download Docker [here](https://www.docker.com/products/docker-desktop).
* You have installed version 1.20 of Go. If not, you can download Go [here](https://golang.org/dl/).
* Your system runs a Linux distribution.

## Running Redis with Docker
To start a Redis instance, run the following command:

```shell
sudo docker run -it -p6379:6379 redis
```

## Building and Running BBQueue
From the root directory of the project, run either of the following commands:

```
go build .
```
or
```
go run .
```
Make sure Redis is running before starting the service.

## Testing
The main function currently runs stress tests on all functionalities of the application.

## Features and Limitations
BBQueue currently only supports a single queue. Here is a list of future improvements that we're planning to implement:

* Refactor cognitive complexity
* Create server struct that loads and holds configs (also change `setupEndpoits` to be a method of server)
* Add support for multiple queues
* Improve `GetAndDeleteExpiredKeys` performance
* Improve redundancy
* Add retry count on processed messages and possibly create a dead message queue for messages over a set limit of retries
* Create/improve logging/stats
* Create Dockerfile that spins up Redis


## API Documentation
API documentation can be found here when the server is running.

## Contributing to BBQueue
We welcome contributions! Please feel free to submit pull requests.

## License
This project uses the MIT license.

