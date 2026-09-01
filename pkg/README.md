###### September 1, 2026: interlace-go pkg readme v6

# PKG

```text
pkg
├── contractor (1)
├── cowriter (2)
├── extractor (2)
├── randommer (2)
├── runner (21)
├── storekeeper (14)
├── tester (1)
└── versioner (2)
```

## Contents

### 1. Contractor

The **Contractor** provides **interlace-go** applications with a consistent workspace and access to several library methods, with a focus on **evet** and **entity** processes. It uses the **Refiner** to orchestrate several roles: the **Manager** sets up and upholds the application environment; the **Dispenser** establishes and maintains the workspace structure; the **Interlacer** encodes the property binary files; the **Cataloguer** encodes the unique-text binary hash tables; the **Enumerator** encodes the controlled-vocabulary binary files; and the **Cowriter** writes the necessary artifacts.

- [Refiner](./contractor/refiner.go)

### 2. Cowriter

A workflow regularly includes at least one **Cowriter**. A **Cowriter** can open a file handler for writing, handle compression, keep track of the file state, cache formatted writes (e.g., tuples) into the associated buffer, flush the buffer to disk, count the number of tuples written to disk, close the file handler, sort the file using the system's command sort, remove the file from the file system, and reset for next usage.

- [Cowriter](./cowriter/cowriter.go)
- [Cowriter Test](./cowriter/cowriter_test.go)

### 3. Extractor

The **Extractor** provides methods and attributes for reading, buffering, and processing big files efficiently.

- [Extractor](./extractor/extractor.go)
- [Extractor Test](./extractor/extractor_test.go)

### 4. Randommer

This is a simple provider of functions for generating random numbers, text, email addresses, phone numbers, DNA sequences, and other random data.

- [Randommer](./randommer/randommer.go)
- [Randommer Test](./randommer/randommer_test.go)

### 5. Runner

The **Runner** controls the tools to configure and run an application. It has special goals for the **Dispenser** and the **Manager**. The **Manager** encapsulates configuration, flags, logging, and reporting mechanisms to drive the application state and behavior. The **Dispenser’s** secures a consistent file system that prevents corruption and guarantees the reproduction of data processing **events** and **entity** datasets across distributed systems. The **Dispenser** directory structure is designed to organize data related to **agents, entities, and events (AEE)** in a hierarchical structure. It follows a specific naming convention to ensure data organization is unique and clear. The AEE structure is conceived to be comprehensible and straightforward with versioned namespaces based on ontology concepts and meaningful asset names. The **agent** directory handles agent configuration and pipeline execution scripts. The **event** directory captures workflow logs and Bash scripts generated at runtime. The **entity** directory stores text catalogs, controlled vocabulary enums, and data property binary files. This system journals all processing events. The outcome compounds and establishes a knowledge hub for downstream analysis.


- [Dispenser](./runner/dispenser.go)
- [Dispenser Test](./runner/dispenser_test.go)
- [Downloader](./runner/downloader.go)
- [Downloader Test](./runner/downloader_test.go)
- [Tramer](./runner/framer.go)
- [Framer Test](./runner/framer_test.go)
- [Helper](./runner/helper.go)
- [Helper Test](./runner/helper_test.go)
- [Manager](./runner/manager.go)
- [Manager Test](./runner/manager_test.go)
- [Masker](./runner/masker.go)
- [Masker Test](./runner/masker_test.go)
- [Scanner](./runner/scanner.go)
- [Scanner Test](./runner/scanner_test.go)
- [Slicer](./runner/slicer.go)
- [Slicer Test](./runner/slicer_test.go)
- [Stacker](./runner/stacker.go)
- [Stacker Test](./runner/stacker_test.go)
- [Validator](./runner/validator.go)
- [Validator Test](./runner/validator_test.go)
- [Wordfinder](./runner/wordfinder.go)

### 6. Storekeeper

Data structures, methods, and functions for storing and searching data properties, controlled-vocabulary enumerations, and unique-text catalogs are delivered along with streaming capabilities and custom ID generators. **Entity** data is stored vertically partitioned in uncurroptible binary files with magic signatures for verification of version and integrity at read time.

- [Cataloguer](./storekeeper/cataloguer.go)
- [Cataloguer Test](./storekeeper/cataloguer_test.go)
- [Container](./storekeeper/container.go)
- [Container Test](./storekeeper/container_test.go)
- [Encoder](./storekeeper/encoder.go)
- [Encoder Test](./storekeeper/encoder_test.go)
- [Enumerator](./storekeeper/enumerator.go)
- [Enumerator Test](./storekeeper/enumerator_test.go)
- [Identifier](./storekeeper/identifier.go)
- [Identifier Test](./storekeeper/identifier_test.go)
- [Interlacer](./storekeeper/interlacer.go)
- [Interlacer Test](./storekeeper/interlacer_test.go)
- [Streamer](./storekeeper/streamer.go)
- [Streamer Test](./storekeeper/streamer_test.go)

### 7. Tester

Tests require location of the absolute paths. **Tester** functions help find the Git repository given its project and repository name.

- [Tester](./tester/tester.go)

### 8. Versioner

**Versioner** stores the current library version in ISO 8601 format.

- [Version](./versioner/version.go)
- [Version Test](./versioner/version_test.go)
