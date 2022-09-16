package modules

func FindIPBlocks(asn string) string {
	output := RequestFunc("https://www.radb.net/query?advanced_query=1&keywords="+asn+"&-T+option=&ip_option=&-i=1&-i+option=origin", "www.radb.net", 5000)
	if len(output) > 0 {
		return output[2]
	}

	return ""
}
