func isTrustedHost(host string) bool {
	// Define a list of trusted hosts
	trustedHosts := []string{"localhost", "mycorp.com"}
	for _, trustedHost := range trustedHosts {
		if host == trustedHost {
			return true
		}
	}
	return false
}

func downloadAndResize(tenantID, fileID, fileSize string) error {
	urlStr := fmt.Sprintf("http://%s.%s/storage/%s.json", tenantID, baseHost, fileID)
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		panic(err)
	}

	if !isTrustedHost(parsedURL.Hostname()) {
		return fmt.Errorf("untrusted host: %s", parsedURL.Hostname())
	}

	// ...
	// ...
}
