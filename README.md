# s3backup

Shorten your strings using common abbreviations.

[![build status](https://travis-ci.org/dnnrly/s3backup.svg?branch=master)](https://travis-ci.org/dnnrly/s3backup)
[![codecov](https://codecov.io/gh/dnnrly/s3backup/branch/master/graph/badge.svg)](https://codecov.io/gh/dnnrly/s3backup)
[![godoc](https://godoc.org/github.com/dnnrly/s3backup?status.svg)](http://godoc.org/github.com/dnnrly/s3backup)
[![report card](https://goreportcard.com/badge/github.com/dnnrly/s3backup)](https://goreportcard.com/report/github.com/dnnrly/s3backup)

## Motivation

This tool has been developed so that I can conveniently backup all of my personal photos to an AWS S3 bucket.

But why develop this? Aren't there other tools that can solve your problem?

Of course, but this is more fun.

## Installation

```bash
git clone https://github.com/dnnrly/
cd s3backup
make build
```

## Usage

```
$ s3backup --help
This too backs up your files to S3 so that you can have them in
the cloud. It will scan the location(s) that you specify and
attempt rudimentary de-duplication.

```

Examples:
```
$ s3backup
```

## Code of Conduct
This project adheres to the Contributor Covenant [code of conduct](CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code.

## Contributing
Pull requests are welcome. See the [contributing guide](CONTRIBUTING.md) for more details.

Please make sure to update tests as appropriate.

## License
[Apache 2](https://choosealicense.com/licenses/apache-2.0/)
