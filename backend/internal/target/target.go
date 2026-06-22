package target

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
)

func Normalize(raw string, requestedPort int) (string, int, error) {
	value := strings.TrimSpace(strings.ToLower(raw))
	if value == "" {
		return "", 0, fmt.Errorf("domain is required")
	}

	if !strings.Contains(value, "://") {
		value = "https://" + value
	}
	parsed, err := url.Parse(value)
	if err != nil {
		return "", 0, fmt.Errorf("invalid domain")
	}

	host := strings.TrimSuffix(parsed.Hostname(), ".")
	if host == "" || strings.ContainsAny(host, " /") {
		return "", 0, fmt.Errorf("invalid domain")
	}
	if net.ParseIP(host) == nil {
		if len(host) > 253 || !strings.Contains(host, ".") {
			return "", 0, fmt.Errorf("enter a fully qualified domain name")
		}
		for _, label := range strings.Split(host, ".") {
			if label == "" || len(label) > 63 || strings.HasPrefix(label, "-") || strings.HasSuffix(label, "-") {
				return "", 0, fmt.Errorf("invalid domain label")
			}
		}
	}

	port := requestedPort
	if port == 0 && parsed.Port() != "" {
		port, err = strconv.Atoi(parsed.Port())
		if err != nil {
			return "", 0, fmt.Errorf("invalid port")
		}
	}
	if port == 0 {
		port = 443
	}
	if port < 1 || port > 65535 {
		return "", 0, fmt.Errorf("port must be between 1 and 65535")
	}

	return host, port, nil
}
