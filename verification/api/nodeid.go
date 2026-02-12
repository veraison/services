// Copyright 2025 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/veraison/services/log"
)

const (
	nodeIDLength   = 6 // bytes, as required by UUID v1
	nodeIDFileName = "veraison-node-id"
)

// getNodeID returns a unique identifier for this node. It tries multiple methods
// in order of preference:
// 1. Read from a persistent node ID file (if exists)
// 2. Use MAC address from the first available non-loopback interface
// 3. Use machine-id if available (fallback for systemd systems)
// 4. Generate a random node ID and persist it
func getNodeID() ([]byte, error) {
	// Try reading from our persistent node ID file
	if id, err := readPersistedNodeID(); err == nil {
		log.Debug("using persisted node ID")
		return id, nil
	}

	// Try getting MAC address
	if id, err := getMACBasedID(); err == nil {
		log.Debug("using MAC-based node ID")
		if err := persistNodeID(id); err != nil {
			log.Warnf("failed to persist node ID: %v", err)
		}
		return id, nil
	}

	// Try machine-id as fallback for systemd systems
	if id, err := getMachineID(); err == nil {
		log.Debug("using machine-id based node ID")
		if err := persistNodeID(id); err != nil {
			log.Warnf("failed to persist node ID: %v", err)
		}
		return id, nil
	}

	// Generate random ID as last resort
	id, err := generateRandomNodeID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate random node ID: %v", err)
	}

	log.Debug("using generated random node ID")
	if err := persistNodeID(id); err != nil {
		log.Warnf("failed to persist node ID: %v", err)
	}

	return id, nil
}

// readPersistedNodeID attempts to read the node ID from a persistent file
func readPersistedNodeID() ([]byte, error) {
	dir := getNodeIDDir()
	path := filepath.Join(dir, nodeIDFileName)

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if len(data) != nodeIDLength*2 { // hex encoded
		return nil, fmt.Errorf("invalid node ID length in file")
	}

	return hex.DecodeString(string(data))
}

// persistNodeID saves the node ID to a persistent file
func persistNodeID(id []byte) error {
	dir := getNodeIDDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	path := filepath.Join(dir, nodeIDFileName)
	return ioutil.WriteFile(path, []byte(hex.EncodeToString(id)), 0644)
}

// getMACBasedID returns a node ID based on the MAC address of the first
// available non-loopback interface
func getMACBasedID() ([]byte, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for _, iface := range ifaces {
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		if iface.Flags&net.FlagUp == 0 {
			continue
		}
		if len(iface.HardwareAddr) < nodeIDLength {
			continue
		}
		return iface.HardwareAddr[:nodeIDLength], nil
	}

	return nil, fmt.Errorf("no suitable network interface found")
}

// getMachineID attempts to read the systemd machine-id
func getMachineID() ([]byte, error) {
	files := []string{"/etc/machine-id", "/var/lib/dbus/machine-id"}
	var id string

	for _, file := range files {
		if data, err := ioutil.ReadFile(file); err == nil {
			id = strings.TrimSpace(string(data))
			break
		}
	}

	if id == "" {
		return nil, fmt.Errorf("no machine-id found")
	}

	// Use first 6 bytes of machine-id hash
	decoded, err := hex.DecodeString(id)
	if err != nil {
		return nil, fmt.Errorf("invalid machine-id format: %v", err)
	}

	return decoded[:nodeIDLength], nil
}

// generateRandomNodeID creates a random node ID
func generateRandomNodeID() ([]byte, error) {
	id := make([]byte, nodeIDLength)
	_, err := rand.Read(id)
	if err != nil {
		return nil, err
	}
	// Set multicast bit as per RFC 4122
	id[0] |= 0x01
	return id, nil
}

// getNodeIDDir returns the directory where the node ID file should be stored
func getNodeIDDir() string {
	if dir := os.Getenv("VERAISON_NODE_ID_DIR"); dir != "" {
		return dir
	}
	return "/var/lib/veraison"
}