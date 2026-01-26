# Protocol & Packet Documentation

This directory contains specifications and technical details about the DOR protocol messages and packet structures.

## Protocol Overview

The DOR protocol uses multi-layered onion encryption where each relay processes and strips one encryption layer from the packet. This document set provides the technical details needed for implementation and analysis.

## Key Components

### Cryptographic Seeding

![Crypto Seeding Operation](./crypto_operation.svg)

Crypto seeding is the process of deriving encryption keys for each layer. Each relay receives a seeded key that allows it to:

1. Decrypt its layer of the onion
2. Generate the next layer's key for downstream relays
3. Process the inner payload

### Onion Packet Structure

![Onion Packet Datagram](./onion_builder.svg)

The onion packet datagram illustrates:

- Header structure and metadata
- Layered encryption organization
- Payload location within the packet
- Forward direction and processing order

### Onion Sequence & Processing

![Onion Processing Sequence](./onion_sequence.d2)

The onion sequence diagram shows:

- Step-by-step processing of the onion packet through relays
- Layer decryption and unwrapping at each relay
- Key derivation and forwarding operations
- Message delivery to the final recipient

### Size Calculations

**[Average Size Calculation](./OnionPacket_AverageSizeCalculation.md)** - Detailed analysis of:

- Base packet overhead
- Per-layer encryption overhead
- Header and metadata sizes
- Typical message sizes under various configurations

This analysis helps in:

- Bandwidth estimation
- Packet fragmentation planning
- Network load modeling
