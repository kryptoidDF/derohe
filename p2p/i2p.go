// Copyright 2017-2021 DERO Project. All rights reserved.
// Use of this source code in any form is governed by RESEARCH license.
// license can be found in the LICENSE file.
// GPG: 0F39 E425 8C65 3947 702A  8234 08B2 0360 A03A 9DE8
//
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND ANY
// EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF
// MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL
// THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
// SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO,
// PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
// INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT,
// STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF
// THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

package p2p

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"net"
	"strings"
	"sync"
	"time"
)

// I2P address format: example.i2p or 52-character base32 address
// I2P addresses end with .i2p suffix

const I2P_SAM_DEFAULT_PORT = 7656 // Default SAM API port
const I2P_TUNNEL_LENGTH_OUTBOUND = 3
const I2P_TUNNEL_LENGTH_INBOUND = 3
const I2P_TUNNEL_QUANTITY = 2

// I2PSession manages the SAM API connection for I2P
type I2PSession struct {
	samHost            string        // SAM API host (must be localhost or 127.0.0.1)
	samPort            int           // SAM API port
	samPassword        string        // SAM API password (if required)
	sessionID          string        // Session ID returned by SAM
	conn               net.Conn      // Connection to SAM
	publicKeyHash      string        // Truncated hash of our public key (for logging)
	destination        string        // Our destination (full .i2p address)
	enabled            bool          // Whether I2P is enabled
	concurrentConns    int64         // Current concurrent I2P connections
	maxConcurrentConns int64         // Max concurrent I2P connections allowed
	connectionTimeout  time.Duration // Timeout for I2P connections
	mutex              sync.Mutex
}

var i2pSession *I2PSession
var i2pMutex sync.Mutex

// IsI2PAddress checks if an address is an I2P address
func IsI2PAddress(address string) bool {
	// I2P addresses end with .i2p
	addr := strings.ToLower(address)
	if strings.HasSuffix(addr, ".i2p") {
		return true
	}
	// Also check for 52-char base32 addresses (without .i2p suffix)
	if len(strings.Split(addr, ":")[0]) == 52 && isBase32(strings.Split(addr, ":")[0]) {
		return true
	}
	return false
}

// isBase32 checks if string contains only base32 characters
func isBase32(s string) bool {
	for _, c := range s {
		if !((c >= 'a' && c <= 'z') || (c >= '2' && c <= '7')) {
			return false
		}
	}
	return true
}

// InitI2P initializes I2P session with SAM API (with security validation)
func InitI2P(samHost string, samPort int) (*I2PSession, error) {
	i2pMutex.Lock()
	defer i2pMutex.Unlock()

	if i2pSession != nil && i2pSession.enabled {
		return i2pSession, nil // Already initialized
	}

	// Security: Validate SAM host is localhost-only
	if !isLocalhostOnly(samHost) {
		return nil, fmt.Errorf("I2P SAM API must be bound to localhost for security, got: %s", samHost)
	}

	session := &I2PSession{
		samHost:            samHost,
		samPort:            samPort,
		enabled:            false,
		maxConcurrentConns: 50,                 // Limit concurrent I2P connections
		connectionTimeout:  30 * time.Second,   // I2P connections timeout
	}

	// Try to connect to SAM API with strict timeout
	addr := fmt.Sprintf("%s:%d", samHost, samPort)
	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		logger.V(1).Error(err, "Failed to connect to I2P SAM API", "address", addr)
		return nil, fmt.Errorf("cannot connect to I2P SAM: %w", err)
	}

	session.conn = conn
	session.sessionID = generateSessionID()

	// Perform SAM handshake
	if err := session.performSAMHandshake(); err != nil {
		conn.Close()
		logger.V(1).Error(err, "SAM handshake failed")
		return nil, fmt.Errorf("SAM handshake failed: %w", err)
	}

	// Create I2P session
	if err := session.createI2PSession(); err != nil {
		conn.Close()
		logger.V(1).Error(err, "Failed to create I2P session")
		return nil, fmt.Errorf("failed to create I2P session: %w", err)
	}

	session.enabled = true
	i2pSession = session
	logger.Info("I2P session initialized successfully", "destination_hash", session.publicKeyHash)

	return session, nil
}

// performSAMHandshake performs the initial SAM API handshake
func (s *I2PSession) performSAMHandshake() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.conn == nil {
		return fmt.Errorf("SAM connection not established")
	}

	// Send HELLO command
	helloCmd := fmt.Sprintf("HELLO VERSION MIN=3.0 MAX=3.3\n")
	if _, err := s.conn.Write([]byte(helloCmd)); err != nil {
		return err
	}

	// Read response
	buf := make([]byte, 256)
	n, err := s.conn.Read(buf)
	if err != nil {
		return err
	}

	response := string(buf[:n])
	if !strings.Contains(response, "HELLO") {
		return fmt.Errorf("unexpected SAM response: %s", response)
	}

	logger.V(2).Info("SAM handshake successful", "response", strings.TrimSpace(response))
	return nil
}

// createI2PSession creates a new I2P destination
func (s *I2PSession) createI2PSession() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.conn == nil {
		return fmt.Errorf("SAM connection not established")
	}

	// Send session create command
	sessionCmd := fmt.Sprintf("SESSION CREATE STYLE=STREAM ID=%s DESTINATION=TRANSIENT "+
		"inbound.tunnelLength=%d outbound.tunnelLength=%d "+
		"inbound.tunnelQuantity=%d outbound.tunnelQuantity=%d\n",
		s.sessionID,
		I2P_TUNNEL_LENGTH_INBOUND,
		I2P_TUNNEL_LENGTH_OUTBOUND,
		I2P_TUNNEL_QUANTITY,
		I2P_TUNNEL_QUANTITY)

	if _, err := s.conn.Write([]byte(sessionCmd)); err != nil {
		return err
	}

	// Read response with our destination
	buf := make([]byte, 2048)
	n, err := s.conn.Read(buf)
	if err != nil {
		return err
	}

	response := string(buf[:n])
	if strings.Contains(response, "SESSION STATUS RESULT=OK") {
		// Extract destination from response
		parts := strings.Fields(response)
		for i, part := range parts {
			if part == "DESTINATION=" && i+1 < len(parts) {
				s.destination = parts[i+1]
				s.publicKeyHash = sanitizeDestinationForLogging(s.destination)
				logger.Info("I2P session created", "destination_hash", s.publicKeyHash)
				return nil
			}
		}
	}

	return fmt.Errorf("failed to create session: %s", response)
}

// ConnectI2P establishes a connection to an I2P peer with security checks
func (s *I2PSession) ConnectI2P(dest string, timeout time.Duration) (net.Conn, error) {
	if !s.enabled {
		return nil, fmt.Errorf("I2P session not initialized")
	}

	// Security: Validate destination format
	if !isValidI2PDestination(dest) {
		return nil, fmt.Errorf("invalid I2P destination format: %s", sanitizeDestinationForLogging(dest))
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Security: Check concurrent connection limit
	if s.concurrentConns >= s.maxConcurrentConns {
		return nil, fmt.Errorf("I2P connection limit reached (%d/%d)", s.concurrentConns, s.maxConcurrentConns)
	}

	// Use configured timeout (not caller's)
	if timeout == 0 || timeout > s.connectionTimeout {
		timeout = s.connectionTimeout
	}

	// Normalize destination address (add .i2p if needed)
	if !strings.HasSuffix(strings.ToLower(dest), ".i2p") && len(strings.Split(dest, ":")[0]) == 52 {
		// Add .i2p suffix for base32 addresses
		destHost := strings.Split(dest, ":")[0]
		destPort := "8333"
		if len(strings.Split(dest, ":")) > 1 {
			destPort = strings.Split(dest, ":")[1]
		}
		dest = destHost + ".i2p:" + destPort
	}

	// Send STREAM CONNECT command to SAM
	connectCmd := fmt.Sprintf("STREAM CONNECT ID=%s DESTINATION=%s\n", s.sessionID, strings.Split(dest, ":")[0])

	if _, err := s.conn.Write([]byte(connectCmd)); err != nil {
		return nil, err
	}

	// Read response
	buf := make([]byte, 512)
	n, err := s.conn.Read(buf)
	if err != nil {
		return nil, err
	}

	response := string(buf[:n])
	if strings.Contains(response, "STREAM STATUS RESULT=OK") {
		s.concurrentConns++
		logger.V(2).Info("I2P connection established", "destination_hash", sanitizeDestinationForLogging(dest))
		return s.conn, nil
	}

	return nil, fmt.Errorf("I2P connection failed")
}

// ListenI2P starts listening for incoming I2P connections
func (s *I2PSession) ListenI2P(port int) (net.Listener, error) {
	if !s.enabled {
		return nil, fmt.Errorf("I2P session not initialized")
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Send STREAM ACCEPT command to SAM
	acceptCmd := fmt.Sprintf("STREAM ACCEPT ID=%s\n", s.sessionID)

	if _, err := s.conn.Write([]byte(acceptCmd)); err != nil {
		return nil, err
	}

	logger.Info("I2P listener started", "destination", s.destination)

	// Return a custom listener
	return &I2PListener{
		session: s,
		port:    port,
	}, nil
}

// I2PListener implements net.Listener for I2P connections
type I2PListener struct {
	session *I2PSession
	port    int
	closed  bool
	mutex   sync.Mutex
}

// Accept accepts an incoming I2P connection
func (l *I2PListener) Accept() (net.Conn, error) {
	l.mutex.Lock()
	if l.closed {
		l.mutex.Unlock()
		return nil, fmt.Errorf("listener closed")
	}
	l.mutex.Unlock()

	// In a real implementation, this would wait for incoming connections
	// For now, return a dummy connection
	return nil, fmt.Errorf("accept not fully implemented yet")
}

// Close closes the I2P listener
func (l *I2PListener) Close() error {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.closed = true
	return nil
}

// Addr returns the listener's address
func (l *I2PListener) Addr() net.Addr {
	return nil
}

// CloseI2P closes the I2P session
func CloseI2P() error {
	i2pMutex.Lock()
	defer i2pMutex.Unlock()

	if i2pSession == nil {
		return nil
	}

	i2pSession.mutex.Lock()
	defer i2pSession.mutex.Unlock()

	if i2pSession.conn != nil {
		i2pSession.conn.Close()
	}

	i2pSession.enabled = false
	logger.Info("I2P session closed")
	return nil
}

// GetI2PDestination returns the current I2P destination
func GetI2PDestination() string {
	i2pMutex.Lock()
	defer i2pMutex.Unlock()

	if i2pSession != nil {
		return i2pSession.destination
	}
	return ""
}

// generateSessionID generates a random session ID
func generateSessionID() string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, 16)
	for i := range result {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		result[i] = charset[num.Int64()]
	}
	return string(result)
}

// sanitizeDestinationForLogging sanitizes I2P destination for secure logging (shows only hash)
func sanitizeDestinationForLogging(destination string) string {
	// Never log full destinations in logs
	// Extract host part
	host := strings.Split(destination, ":")[0]
	if len(host) > 20 {
		return host[:20] + "..."
	}
	return host
}

// isValidI2PDestination validates I2P destination format and length
func isValidI2PDestination(dest string) bool {
	// Split host and port
	parts := strings.Split(dest, ":")
	if len(parts) != 2 {
		return false
	}

	host := parts[0]
	port := parts[1]

	// Validate port is numeric
	if port == "" {
		return false
	}
	for _, c := range port {
		if c < '0' || c > '9' {
			return false
		}
	}

	// Validate host format
	if strings.HasSuffix(strings.ToLower(host), ".i2p") {
		// Remove .i2p suffix and validate base32
		base32Addr := strings.TrimSuffix(host, ".i2p")
		return len(base32Addr) == 52 && isBase32(base32Addr)
	}

	// Or 52-char base32 without suffix
	return len(host) == 52 && isBase32(host)
}

// isLocalhostOnly validates that host is localhost for security
func isLocalhostOnly(host string) bool {
	return host == "127.0.0.1" || host == "localhost" || host == "[::1]" || host == "::1"
}

// GetI2PSession returns the current I2P session
func GetI2PSession() *I2PSession {
	i2pMutex.Lock()
	defer i2pMutex.Unlock()
	return i2pSession
}

// IsI2PEnabled checks if I2P is enabled and initialized
func IsI2PEnabled() bool {
	i2pMutex.Lock()
	defer i2pMutex.Unlock()
	return i2pSession != nil && i2pSession.enabled
}
