// Copyright 2017-2021 DERO Project. All rights reserved.
// Use of this source code in any form is governed by RESEARCH license.
// license can be found in the LICENSE file.

package config

// I2P seed nodes for mainnet
// These nodes operate both IP and I2P connectivity
var Mainnet_I2P_seed_nodes = []string{
	// To be populated with I2P nodes that wish to serve as seed nodes
	// Format: "base32address.i2p:40401" or "base32address:40401"
	// Example: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.i2p:40401",
}

// I2P seed nodes for testnet
var Testnet_I2P_seed_nodes = []string{
	// Testnet I2P nodes
	// Format: "base32address.i2p:40401"
	// Example: "testnetnode1234567890abcdefghijklmnopqrstuvwxyz12345.i2p:40401",
}

// I2P Configuration defaults
const (
	I2P_DEFAULT_ENABLED        = false       // Enable I2P by default (can be overridden via ENABLE_I2P env var)
	I2P_SAM_DEFAULT_HOST       = "127.0.0.1" // Default SAM API host
	I2P_SAM_DEFAULT_PORT       = 7656        // Default SAM API port
	I2P_TUNNEL_LENGTH_INBOUND  = 3           // Inbound tunnel length (privacy vs speed tradeoff)
	I2P_TUNNEL_LENGTH_OUTBOUND = 3           // Outbound tunnel length
	I2P_TUNNEL_QUANTITY        = 2           // Number of parallel tunnels
	I2P_CONNECTION_TIMEOUT     = 30          // Seconds to wait for I2P connection
	I2P_MAINTENANCE_INTERVAL   = 5           // Seconds between seed node maintenance checks
)
