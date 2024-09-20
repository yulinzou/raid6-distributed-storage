# RAID-6 Distributed Storage System
## Overview
This project implements a simplified RAID-6 Distributed Storage System in Go. The system stripes data across multiple storage nodes and adds fault tolerance by calculating two types of parity (P-parity and Q-parity), which allows recovery from up to two simultaneous node failures.

The project simulates storage nodes, generates synthetic data, and supports node failure simulation and data recovery using the parity blocks.

## Features
* Data Striping: Divides data into blocks and distributes it across storage nodes.
* Dual Parity (P and Q): Implements P-parity using XOR and includes a placeholder for Q-parity using Reed-Solomon encoding.
* Fault Tolerance: Supports recovery from up to two node failures.
* File Content update: Update content of file given the name and new content of the file.
* Disk Persistence: Read/write data blocks on disk, persistent data.
* Flexible Disk Number: Support more than 6+2 nodes to n+2 nodes.

## Experiments
We have conducted several experiments on the system.
* Functionality check with each feature:
    - File write/read
    - Node recovery
    - Content update
    - Disk persistence
    - Stress test on large files
* Analytical experiments on system performance:
    - Computation latency
    - I/O latency
    - Disk number impact
