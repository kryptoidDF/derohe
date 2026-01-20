# DERO I2P Integration Guide

## Overview

The DERO daemon now supports I2P (Invisible Internet Project) connectivity, allowing nodes to connect to both regular IP-based peers and anonymous I2P-based peers. This provides enhanced privacy and censorship resistance capabilities.

## Architecture

### Components

1. **i2p.go** - Core I2P module handling SAM API communication
2. **controller.go** - Integration with P2P engine for I2P connections
3. **Connection Management** - Unified connection pool for both regular and I2P connections

### Key Features

- **Dual-stack Connectivity**: Connect to both regular IPv4/IPv6 nodes and I2P nodes
- **SAM API Integration**: Uses I2P's Simple Anonymous Messaging (SAM) API for communication
- **Automatic Fallback**: Gracefully handles I2P unavailability
- **Peer Management**: Tracks I2P peers separately in the peer list
- **Connection Pooling**: Manages I2P connections through the same connection pool

## Setup Instructions

### Prerequisites

1. **I2P Router** - Download and install from [i2p.net](https://i2p.net)
2. **SAM API Enabled** - Ensure SAM is enabled in your I2P configuration
   - Edit: `~/.i2p/clients.config`
   - Add/ensure: `clientApp.0=net.i2p.sam.SAMBridge`
3. **Network Access** - SAM API must be accessible (typically localhost:7656)

### Configuration

#### Environment Variables

```bash
# Enable I2P support
export ENABLE_I2P=1

# I2P SAM API host and port (optional, defaults to localhost:7656)
export I2P_SAM_HOST=127.0.0.1
export I2P_SAM_PORT=7656
```

#### Starting DERO Daemon with I2P

```bash
# Testnet with I2P enabled
ENABLE_I2P=1 ./derod --testnet

# Mainnet with custom SAM port
ENABLE_I2P=1 I2P_SAM_PORT=7656 ./derod
```

### I2P Seed Nodes

I2P seed node configuration will be added to `config/seed_nodes.go`:

```go
// I2P seed nodes for testnet
var Testnet_I2P_seed_nodes = []string{
    "example.i2p:40401",
}

// I2P seed nodes for mainnet
var Mainnet_I2P_seed_nodes = []string{
    "example.i2p:40401",
}
```

## Usage

### I2P Address Format

DERO supports two I2P address formats:

1. **Base32 address** (52 characters):
   ```
   aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.i2p:40401
   ```

2. **Destination hash** with `.i2p` suffix:
   ```
   [52-char-base32].i2p:40401
   ```

### Adding I2P Peers

```bash
# Connect to a specific I2P node (via daemon command)
# The daemon will automatically detect and route to I2P

# Example peer discovery message
# Peers can advertise their I2P addresses in addition to IP addresses
```

### Monitoring I2P Connections

Check daemon logs for I2P connection status:

```
I2P session initialized successfully
I2P connection established to: example.i2p:40401
I2P seed node maintenance tick
```

## API Reference

### Public Functions

#### `InitI2P(samHost string, samPort int) (*I2PSession, error)`
Initialize I2P session with SAM API.

**Parameters:**
- `samHost`: SAM API host (usually "127.0.0.1")
- `samPort`: SAM API port (usually 7656)

**Returns:**
- `*I2PSession`: Session handle
- `error`: Initialization error if any

#### `IsI2PAddress(address string) bool`
Check if an address is a valid I2P address.

**Parameters:**
- `address`: Address to validate (e.g., "example.i2p:40401")

**Returns:**
- `true` if valid I2P address format

#### `IsI2PEnabled() bool`
Check if I2P is initialized and enabled.

#### `GetI2PDestination() string`
Get the daemon's I2P destination address.

#### `CloseI2P() error`
Gracefully close I2P session.

### Connection Functions

```go
// Establish outgoing I2P connection
func (s *I2PSession) ConnectI2P(dest string, timeout time.Duration) (net.Conn, error)

// Listen for incoming I2P connections
func (s *I2PSession) ListenI2P(port int) (net.Listener, error)
```

## Network Behavior

### Connection Priority

1. **Regular connections** - Established using KCP over UDP
2. **I2P connections** - Established through SAM API
3. **Automatic routing** - Determined by address format

### Peer Discovery

- **IP peers** - Discovered through regular P2P protocol
- **I2P peers** - Discovered through I2P network or configured seed nodes
- **Mixed network** - Nodes can be both IP and I2P peers

### Bandwidth Considerations

I2P connections may have higher latency than direct IP connections. The daemon automatically:
- Applies appropriate timeouts for I2P connections
- Manages connection pool to prevent saturation
- Implements separate backoff strategies for I2P peers

## Security Considerations

### Privacy Benefits

- I2P connections are routed through multiple anonymous hops
- Remote peers cannot determine your IP address
- Network-level ISP tracking becomes significantly harder

### Security Model

- **Connection Encryption**: I2P + TLS (consistent with regular connections)
- **Peer Validation**: Same block validation regardless of connection type
- **Consensus Rules**: No different treatment for I2P-originated transactions

## Troubleshooting

### I2P Connection Failures

**Problem**: "Failed to connect to I2P SAM API"

**Solutions**:
1. Verify I2P router is running: `ps aux | grep i2p`
2. Check SAM is enabled in I2P config
3. Verify SAM port: `netstat -tuln | grep 7656`
4. Check firewall rules

### No I2P Peers

**Problem**: No connections being made to I2P nodes

**Solutions**:
1. Verify I2P network is synchronized
2. Check peer list for I2P entries: `peers list` (daemon command)
3. Ensure seed nodes are reachable
4. Check logs for connection attempts

### I2P Session Errors

**Problem**: SAM protocol errors or handshake failures

**Solutions**:
1. Check I2P router logs for errors
2. Verify I2P version compatibility (SAM 3.0+)
3. Restart I2P router
4. Rebuild I2P if necessary

## Performance Optimization

### Tuning I2P Parameters

Edit `i2p.go` constants:

```go
const I2P_TUNNEL_LENGTH_OUTBOUND = 3  // Increase for privacy, decrease for speed
const I2P_TUNNEL_LENGTH_INBOUND = 3
const I2P_TUNNEL_QUANTITY = 2         // More tunnels = better throughput
```

### Monitoring I2P

Environment variables for tuning:

```bash
# Increase logging for I2P connections
export DERO_LOG_LEVEL=2  # More verbose

# Monitor connection metrics
./derod --testnet  # Check metrics endpoint
```

## Future Enhancements

- [ ] I2P node address advertising
- [ ] Separate I2P seed node management UI
- [ ] I2P-only operation mode
- [ ] I2P tunnel quality metrics
- [ ] Automatic I2P router discovery
- [ ] Multi-hop I2P mixing strategies

## References

- [I2P Project](https://i2p.net)
- [SAM Bridge Specification](https://geti2p.net/en/docs/api/samv3)
- [DERO P2P Protocol](./README.md)

## Support

For issues or questions:
1. Check daemon logs for I2P errors
2. Verify I2P router is functioning correctly
3. File an issue on GitHub with logs
4. Test with testnet before mainnet deployment
