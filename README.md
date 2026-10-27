###### October 27, 2026: interlace-go readme v28

# Welcome to Quantome’s Interlace Go Library

[![Build Status](https://github.com/sas-quantome/interlace-go/actions/workflows/release.yml/badge.svg)](https://github.com/sas-quantome/interlace-go/actions/workflows/release.yml) [![Latest Release](https://img.shields.io/github/v/release/sas-quantome/interlace-go?include_prereleases&sort=semver)](https://github.com/sas-quantome/interlace-go/releases) [![Go Version](https://img.shields.io/github/go-mod/go-version/sas-quantome/interlace-go)](https://github.com/sas-quantome/interlace-go/blob/main/go.mod)

This library helps innovators refine data and consolidate it into a ready state for any downstream tool, like pluripotent stem cells. It ingests and synthesizes disparate datasets into immutable, vertically partitioned, strongly typed binary arrays. This parallel substrate enables seamless, additive schema evolution while maintaining absolute data integrity and agile streaming fluidity.

The **interlace-go** packages are designed to build robust command-line applications for high-speed information processing and querying, piping deterministic data directly into downstream tools such as LLMs, relational databases, edge computing systems, bioinformatics pipelines, and standard Linux workflows.

![interlace-go](assets/interlace-go.png)

## Contents

1) [Introduction](#1-introduction)
2) [Installation](#2-installation)
3) [Test](#3-test)
4) [Usage](#4-usage)
5) [License](#5-license--legal-notice)
6) [Support and Commercial SLAs](#6-support-and-commercial-slas)

### 1. Introduction

A database is a transient tool and can be treated as a disposable secondary store. True independence from vendor lock-in requires data products built for additive schema evolution, free of infrastructure bloat, and liberated from archaic standards.

#### 1.1 Overall Architecture: Extensible Columnar Storage

The **interlace-go** library stores data by column (e.g., names in one file, SSNs in another) rather than by row (e.g., the whole JSON record in one file), this is a **Columnar Storage** (or Column-Oriented) format. Because new columns can be added at will without rewriting the old ones, it is **Extensible** or **Dynamic**.

#### 1.2 Data Structure: Structure of Arrays (SoA) / Parallel Arrays

By maintaining separate arrays for each attribute and using the index (e.g., index 0 in the key array matches index 0 in the other arrays) to reconstruct a record, we use the **Structure of Arrays (SoA)** pattern. The arrays themselves are called **Parallel Arrays**.

#### 1.3 Parsing Process: Record Shredding / Vertical Partitioning

The library splits nested records (e.g., a wide JSON document) into individual files for each attribute (or column). This is officially known as **Data Shredding**, a term popularized by Google's Dremel paper, which led to Apache Parquet. At a database architecture level, splitting a table into individual column files is called **Vertical Partitioning**.

#### 1.4 Data Evolution: Additive Schema Evolution

The ability to discover new attributes later and instantly have a new column available without rebuilding the previous files is called **Additive Schema Evolution**. The interlace-go system uses a **late-binding schema**, meaning it doesn't strictly enforce the exact shape of the "table" until it reads the folder.

#### 1.5 Organization: Directory-as-a-Table (Data Lake Pattern)

The **interlace-go** approach groups array files inside a single folder and treats that folder as a unique data asset. This is a standard **Data Lake pattern**. It is often referred to as the **Directory-as-a-Table** concept, where the folder path designates the logical table or **Interlace Asset**, and several assets represent an underlying **Interlace Entity**.

#### 1.6 Documentation: Deep Dive Inside Interlace Go

↳ [pkg][PKG]

### 2. Installation

Create your codebase, initialize Git, and add the **interlace-go** package as a dependency. We organize our repositories as follows, but it is not a requirement:

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

Move into `YourGoCodeBaseName` directory, turn it into a Git repository, and initialize your Go modules where `YourGoModuleName` may look like: `github.com/my-project/my-interlace`.

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

To test **interlace-go**, you must first clone the **interlace-go** repository and then execute the Makefile.

```shell
git clone https://github.com/sas-quantome/interlace-go.git
```

The above command should create the `interlace-go` directory in your local machine. The Makefile inside provides the following targets for testing:

```text
test                           Run tests
coverage                       Run coverage report
vet                            Run code issues detection
clean                          Remove the bin folder and test files
```

Start with `test` and finish with `clean`, for example:

```shell
cd interlace-go
make test
```

### 4. Usage

The following [Quantome's](https://github.com/sas-quantome) GitHub repositories demonstrate how to use the **interlace-go** library.

#### 4.1 [interlace-ex](https://github.com/sas-quantome/interlace-ex)

Learn how to code and build a data stack with the **interlace-ex** command-line examples. This open-source educational repository guides you through consolidating human gene data, including products, interactants, genetic variants, and diseases, as a hands-on data refinement exercise. 15 Go programs are offered as starting models for creating robust datasets. The ensemble of these Go programs creates one **Interlace Entity**, named *Gene*, and 14 *Gene* **Interlace Assets**.

#### 4.2 [interlace-do](https://github.com/sas-quantome/interlace-do)

Once you have built the *Gene* data stack, you can use the **interlace-do** toolset to learn how to search and retrieve data from any **Interlace Asset**. This repository provides more than 40 command-line examples to help you do your own *Gene* queries.

### 5. License & Legal Notice

This project is licensed under the **Apache License 2.0**.

* **Permissive Reuse:** You are free to modify, distribute, and use this software in your own commercial or private applications.
* **No Lock-In:** Your downstream solutions remain entirely your own.
* **As-Is Basis:** Per Sections 7 and 8 of the Apache 2.0 License, this software is provided strictly on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND. 

For full details, please review the [LICENSE](LICENSE) and [NOTICE](NOTICE) files in the root of this repository.

### 6. Support and Commercial SLAs

This repository is an **open-source transparency initiative** and does not come with a dedicated maintenance or support SLA.

Bug reports and feature suggestions can be submitted via GitHub Issues.

If your organization requires guaranteed technical support, custom feature development, or integration assistance, please contact your account manager or reach out via [www.sas-quantome.com](https://www.sas-quantome.com).

***
*&copy; 2026, Société par Actions Simplifiée Quantome, Paris 75014, France. All rights reserved.*

[PKG]: https://github.com/sas-quantome/interlace-go/tree/main/pkg/#PKG
