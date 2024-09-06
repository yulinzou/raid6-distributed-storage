# RAID-6 Distributed Storage System
## Overview
This project implements a simplified RAID-6 Distributed Storage System in Go. The system stripes data across multiple storage nodes and adds fault tolerance by calculating two types of parity (P-parity and Q-parity), which allows recovery from up to two simultaneous node failures.

The project simulates storage nodes, generates synthetic data, and supports node failure simulation and data recovery using the parity blocks.

## Features
* Data Striping: Divides data into blocks and distributes it across storage nodes.
* Dual Parity (P and Q): Implements P-parity using XOR and includes a placeholder for Q-parity using Reed-Solomon encoding.
* Fault Tolerance: Supports recovery from up to two node failures.