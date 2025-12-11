package modules

import "ipmap/config"

// FindIPBlocks queries RADB to find all IP blocks for a given ASN.
// Returns the raw HTML response containing route and route6 entries.
func FindIPBlocks(asn string) string {
	output := RequestFunc("https://www.radb.net/query?advanced_query=1&keywords="+asn+"&-T+option=&ip_option=&-i=1&-i+option=origin", "www.radb.net", config.DefaultAPITimeout)
	if len(output) > 0 {
		return output[2]
	}

	return ""
}
