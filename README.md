[![Build Status](https://travis-ci.org/danfaizer/redis3.svg?branch=master)](https://travis-ci.org/danfaizer/redis3)
[![codecov](https://codecov.io/gh/danfaizer/redis3/branch/master/graph/badge.svg)](https://codecov.io/gh/danfaizer/redis3)
[![Go Report Card](https://goreportcard.com/badge/github.com/danfaizer/redis3)](https://goreportcard.com/report/github.com/danfaizer/redis3)
[![GoDoc](https://godoc.org/github.com/danfaizer/redis3?status.svg)](https://godoc.org/github.com/danfaizer/redis3)
# RediS3
Poor's man HA and distributed key-value storage GO library running on top of AWS S3.

In some projects/PoC you may require some kind of persistence, perhaps accessible from different nodes/processes and in a key-value format.<br>
**RediS3** is a simple **key-value** library that leverages **AWS S3** to provide **HA** and **distributed persistence**.

This library is in early stages and is missing some key features, but you can see what is this about, therefore PR and suggestions are very welcome.

## Features
* [x] HA and distributed (kindly provided by AWS S3)
* [x] Key locking. Soft and Hard consistency
* [x] Store GO objects (GO built in and struct objects)
* [x] Key expiration
* [ ] List keys
* [ ] Read-only client
* [ ] Configurable Exponential Back-Off for AWS calls
* [ ] Client stats/metrics

## Requirements
RediS3 leverages AWS S3 service to persist data. This means that the node running RediS3 requires proper access to AWS S3 service.

There are several ways to provide AWS credentials and proper access level to S3 buckets:

- For testing purposes, run [Moto](https://github.com/spulec/moto) AWS mock locally. Example running Moto in a Docker container:

```bash
docker pull picadoh/motocker
docker run --rm --name s3 -d -e MOTO_SERVICE=s3 -p 5001:5000 -i picadoh/motocker
export AWS_ACCESS_KEY_ID=DUMMYAWSACCESSKEY
export AWS_SECRET_ACCESS_KEY=DUMMYAWSSECRETACCESSKEY

```
- For testing purposes, use [Roly](https://github.com/diasjorge/roly) to use AWS S3 with STS tokens in your local machine.
- If you are running RediS3 on an EC2 instance, attach an [Instance Profile](https://docs.aws.amazon.com/en_en/IAM/latest/UserGuide/id_roles_use_switch-role-ec2.html) with proper permissions to the EC2 instance.
- (Not recommended) use [environment variables](https://docs.aws.amazon.com/cli/latest/userguide/cli-environment.html) to provide proper access credentials.

## Installation

Install:

```shell
go get -u github.com/danfaizer/redis3
```

Import:

```go
import "github.com/danfaizer/redis3"
```

## Quickstart
```go
// Create RediS3 client
client, err := redis3.NewClient(
  &redis3.Options{
    Bucket:             "redis3-database",
    AutoCreateBucket:   true,
    Region:             "eu-west-1",
    Timeout:            1,
    EnforceConsistency: true,
  })
if err != nil {
  panic(err)
}
```
```go
type person struct {
    uuid string
    name string
    age  int
}

p := person{
  uuid: "123e4567-e89b-12d3-a456-426655440000",
  name: "Daniel",
  age:  35,
}

var err error

err = client.Set(p.id, p, 0)
if err != nil {
  panic(err)
}

var b person

_, err = client.Get("123e4567-e89b-12d3-a456-426655440000", &b)
if err != nil {
  panic(err)
}

fmt.Printf("%+v", b)
```
```shell
{uuid:123e4567-e89b-12d3-a456-426655440000 name:Daniel age:35}
```
