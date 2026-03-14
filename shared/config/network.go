package config

import (
	"encoding/xml"
	"os"
)

// NetworkConfig holds network-related configuration parsed from network.xml
type NetworkConfig struct {
	// URL prefixes for the server (comma-separated)
	// Example: "/jellyfin,http://+/:8096"
	URLPrefixes []string `xml:"-" json:"url_prefixes"`

	// HTTPS configuration
	EnableHTTPS   bool   `xml:"EnableHTTPS" json:"enable_https"`
	HTTPSPort     int    `xml:"HTTPSPort" json:"https_port"`
	Certificate   string `xml:"Certificate" json:"certificate"`       // Path to certificate file
	CertificatePW string `xml:"CertificatePassword" json:"certificate_password"` // Certificate password

	// Network bindings
	EnableIPV6             bool     `xml:"EnableIPV6" json:"enable_ipv6"`
	IPV6ToIPv4Enabled      bool     `xml:"IPV6ToIPv4Enabled" json:"ipv6_to_ipv4_enabled"`
	IPV6ToIPv4Prefix       string   `xml:"IPV6ToIPv4Prefix" json:"ipv6_to_ipv4_prefix"`
	PublishedServerURIWithHTTP  string `xml:"PublishedServerUriWithHttp" json:"published_server_uri_with_http"`
	PublishedServerURIWithHTTPS string `xml:"PublishedServerUriWithHttps" json:"published_server_uri_with_https"`

	// Remote access configuration
	EnableRemoteAccess           bool     `xml:"EnableRemoteAccess" json:"enable_remote_access"`
	RemoteIPFilter               []string `xml:"RemoteIPFilter>string" json:"remote_ip_filter"` // Allow/deny list
	EnableRemoteIPFilter         bool     `xml:"EnableRemoteIPFilter" json:"enable_remote_ip_filter"`
	IsRunningInDocker            bool     `xml:"IsRunningInDocker" json:"is_running_in_docker"`

	// Transcoding network settings
	TranscodedCachePath        string `xml:"TranscodedCachePath" json:"transcoded_cache_path"`
	EnableTranscodingThrottling bool   `xml:"EnableTranscodingThrottling" json:"enable_transcoding_throttling"`

	// Local network configuration
	LocalNetworkAddresses         []string `xml:"LocalNetworkAddresses>string" json:"local_network_addresses"`
	RemoteIPFilterRegex           []string `xml:"RemoteIPFilterRegex>string" json:"remote_ip_filter_regex"`
	FilterAssociatedClients       bool     `xml:"FilterAssociatedClients" json:"filter_associated_clients"`
	EnableAutomaticPortForwarding bool     `xml:"EnableAutomaticPortForwarding" json:"enable_automatic_port_forwarding"`

	// HTTP settings
	WriteServerHeader bool   `xml:"WriteServerHeader" json:"write_server_header"`
	RequireHttps      bool   `xml:"RequireHttps" json:"require_https"`
	RedirectToHttps   bool   `xml:"RedirectToHttps" json:"redirect_to_https"`

	// Default values (fallbacks if not set)
	DefaultHTTPSPort int    `json:"default_https_port"` // Typically 8920
	DefaultHTTPPort  int    `json:"default_http_port"`  // Typically 8096
}

// NetworkXML is the XML structure matching network.xml format
type NetworkXML struct {
	XMLName xml.Name `xml:"NetworkConfiguration"`

	// Bind settings
	EnableIPV6               bool   `xml:"EnableIPV6"`
	IPV6ToIPv4Enabled        bool   `xml:"IPV6ToIPv4Enabled"`
	IPV6ToIPv4Prefix         string `xml:"IPV6ToIPv4Prefix"`

	// HTTPS settings
	EnableHTTPS   bool   `xml:"EnableHTTPS"`
	HTTPSPort     int    `xml:"HTTPSPort"`
	Certificate   string `xml:"Certificate"`
	CertificatePW string `xml:"CertificatePassword"`

	// Published server URIs
	PublishedServerUriWithHttp  string `xml:"PublishedServerUriWithHttp"`
	PublishedServerUriWithHttps string `xml:"PublishedServerUriWithHttps"`

	// Remote access
	EnableRemoteAccess    bool     `xml:"EnableRemoteAccess"`
	FilterAssociatedClients         bool     `xml:"FilterAssociatedClients"`
	EnableAutomaticPortForwarding   bool     `xml:"EnableAutomaticPortForwarding"`

	// IP filtering
	RemoteIPFilter           []string `xml:"RemoteIPFilter>string"`
	EnableRemoteIPFilter     bool     `xml:"EnableRemoteIPFilter"`
	RemoteIPFilterRegex      []string `xml:"RemoteIPFilterRegex>string"`

	// Local network
	LocalNetworkAddresses []string `xml:"LocalNetworkAddresses>string"`

	// Transcoding
	TranscodedCachePath        string `xml:"TranscodedCachePath"`
	EnableTranscodingThrottling bool   `xml:"EnableTranscodingThrottling"`

	// HTTP settings
	IsRunningInDocker     bool   `xml:"IsRunningInDocker"`
	WriteServerHeader     bool   `xml:"WriteServerHeader"`
	RequireHttps          bool   `xml:"RequireHttps"`
	RedirectToHttps       bool   `xml:"RedirectToHttps"`
}

// DefaultNetworkConfig returns a NetworkConfig with sensible defaults
func DefaultNetworkConfig() *NetworkConfig {
	return &NetworkConfig{
		DefaultHTTPSPort: 8920,
		DefaultHTTPPort:  8096,
		EnableRemoteAccess: true,
		EnableIPV6:       false,
		WriteServerHeader: true,
		RedirectToHttps:  false,
		RequireHttps:     false,
		IsRunningInDocker: false,
	}
}

// LoadNetworkConfig reads and parses network.xml from the given path
func LoadNetworkConfig(networkXMLPath string) (*NetworkConfig, error) {
	// Check if file exists
	if _, err := os.Stat(networkXMLPath); os.IsNotExist(err) {
		// Return defaults if file doesn't exist
		return DefaultNetworkConfig(), nil
	}

	// Read file contents
	data, err := os.ReadFile(networkXMLPath)
	if err != nil {
		return DefaultNetworkConfig(), err
	}

	// Parse XML
	var networkXML NetworkXML
	if err := xml.Unmarshal(data, &networkXML); err != nil {
		return DefaultNetworkConfig(), err
	}

	// Convert to NetworkConfig
	config := &NetworkConfig{
		EnableIPV6:                networkXML.EnableIPV6,
		IPV6ToIPv4Enabled:         networkXML.IPV6ToIPv4Enabled,
		IPV6ToIPv4Prefix:          networkXML.IPV6ToIPv4Prefix,
		EnableHTTPS:               networkXML.EnableHTTPS,
		HTTPSPort:                 networkXML.HTTPSPort,
		Certificate:               networkXML.Certificate,
		CertificatePW:             networkXML.CertificatePW,
		PublishedServerURIWithHTTP:  networkXML.PublishedServerUriWithHttp,
		PublishedServerURIWithHTTPS: networkXML.PublishedServerUriWithHttps,
		EnableRemoteAccess:        networkXML.EnableRemoteAccess,
		FilterAssociatedClients:   networkXML.FilterAssociatedClients,
		EnableAutomaticPortForwarding: networkXML.EnableAutomaticPortForwarding,
		RemoteIPFilter:            networkXML.RemoteIPFilter,
		EnableRemoteIPFilter:      networkXML.EnableRemoteIPFilter,
		RemoteIPFilterRegex:       networkXML.RemoteIPFilterRegex,
		LocalNetworkAddresses:     networkXML.LocalNetworkAddresses,
		TranscodedCachePath:       networkXML.TranscodedCachePath,
		EnableTranscodingThrottling: networkXML.EnableTranscodingThrottling,
		IsRunningInDocker:         networkXML.IsRunningInDocker,
		WriteServerHeader:         networkXML.WriteServerHeader,
		RequireHttps:              networkXML.RequireHttps,
		RedirectToHttps:           networkXML.RedirectToHttps,
		DefaultHTTPSPort:          8920,
		DefaultHTTPPort:           8096,
	}

	// Set HTTPS port default to 8920 if not specified
	if config.HTTPSPort == 0 {
		config.HTTPSPort = config.DefaultHTTPSPort
	}

	return config, nil
}

// ParseURLPrefixes parses the URL prefixes string into a slice
func (nc *NetworkConfig) ParseURLPrefixes(prefixes string) {
	// Simple comma-separated parsing
	if prefixes == "" {
		return
	}
	// Split by comma and trim spaces
	parts := SplitString(prefixes, ",")
	for _, p := range parts {
		p = TrimString(p)
		if p != "" {
			nc.URLPrefixes = append(nc.URLPrefixes, p)
		}
	}
}

// Helper functions to avoid dependency on strings package
func SplitString(s, sep string) []string {
	var result []string
	start := 0
	for i := 0; i <= len(s)-len(sep); i++ {
		if s[i:i+len(sep)] == sep {
			result = append(result, s[start:i])
			start = i + len(sep)
			i = start - 1
		}
	}
	result = append(result, s[start:])
	return result
}

func TrimString(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}
