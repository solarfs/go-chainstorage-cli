package cmd

import (
	"fmt"
	"io"
	"net/http"
)

// get the header data of response from IPFS gateway
func getHttpHeaderMap(url string) (map[string][]string, error) {
	//url = "https://test-ipfs-gateway.netwarps.com/ipfs/bafybeibawqusphaqfn7c7b5hsr4wgmxud7qfhnmd2ntco7oqjnvmg5njfa"

	response, err := http.Head(url)
	if err != nil {
		fmt.Println("Error:", err)
		return nil, err
	}
	defer response.Body.Close()

	headerMap := make(map[string][]string)
	for key, values := range response.Header {
		headerMap[key] = values
	}

	//// Retrieve and display the response headers
	//for key, values := range headerMap {
	//	for _, value := range values {
	//		fmt.Printf("%s: %s\n", key, value)
	//	}
	//}

	return headerMap, err
}

// download the dag data from IPFS gateway
func getDagData(url string) ([]byte, error) {
	response, err := http.Get(url)
	if err != nil {
		fmt.Println("Error:", err)
		return nil, err
	}
	defer response.Body.Close()

	// Read the response body
	body, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Error:", err)
		return nil, err
	}

	return body, nil
}
