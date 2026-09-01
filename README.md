###### September 1, 2026: interlace-go readme v2

# interlace-go

![interlace-go](assets/interlace-go.png)

## Contents

1) [Introduction](#1-introduction)
2) [Installation](#2-installation)
3) [Test](#3-test)
4) [Usage](#4-usage)
5) [License](#5-license)

### 1. Introduction

The **interlace-go** packages help innovators refine data and consolidate it into a ready state for any downstream tool. It is architected to ingest and synthesize disparate datasets into immutable, vertically partitioned, and strongly typed binary arrays. This parallel substrate permits seamless, additive schema evolution while maintaining absolute data integrity. Most importantly, it enables programming command-line tools for high-speed information processing and querying, piping deterministic data directly into downstream tools such as LLMs, edge computing systems, bioinformatics pipelines, and standard Linux workflows.

### 2. Installation

Create your code base, initialize Git, and add the **interlace-go** package as a dependency. A bare-bone Go repository will look similar to the following:

```text
YourGoCodeBaseName
├── .gitignore
├── LICENSE
├── README.md
├── cmd
│   ├── YourGoAppName1
│   │   └── main.go
│   └── YourGoAppName2
│       └── main.go
├── go.mod
├── go.sum
└── internal
   ├── app
   │   ├── YourGoAppName1
   │   │   ├── YourGoAppName1.go
   │   │   └── YourGoAppName1_test.go
   │   └── YourGoAppName2
   │       ├── YourGoAppName1.go
   │       └── YourGoAppName1_test.go
   └── pkg
       ├── README.md
       ├── YourPkgName
       │   ├── YourPkgName.go
       │   └── YourPkgName_test.go
       └── YourPkg2
           ├── YourPkg2.go
           └── YourPkg2_test.go
```

Move into `YourGoCodeBaseName` directory, turn it into a Git repository, and initialize your Go modules where `YourGoModuleName` may look like: `github.com/YourCodeBase/YourGoCodeBaseName`.

```shell
cd YourGoCodeBaseName
git init
go mod init YourGoModuleName
```

You can reference **interlace-go** in a Go program with import as follows:

```text
import "github.com/sas-quantome/interlace-go"
```

Or, you can explicitly get the package into `YourGoCodeBaseName`:

```shell
cd YourGoCodeBaseName
go get github.com/sas-quantome/interlace-go
```

### 3. Test

To test **interlace-go**, you must first clone the **interlace-go** repository and then run make.

```shell
git clone https://github.com/sas-quantome/interlace-go.git
```

The above command should create the `interlace-go` directory in your local machine. The Makefile inside provides the following targets for testing:

```text
test                           Run tests
coverage                       Run coverage report
vet                            Run code issues detection
clean                          Remove the bin folder and test files leaving the download data files intact
```

Start with `test` and finish with `clean`. Example running make:

```shell
cd interlace-go
make test
```

### 4. Usage

The following examples demonstrate how to use the **interlace-go** packages.

### 5. License

This project is licensed under the Apache License Version 2.0 – see the [LICENSE](LICENSE) file for details.
