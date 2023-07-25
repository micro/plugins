// Package k8sheadlesssvc /*
package k8sheadlesssvc

import (
	"fmt"
	"net"
)

// dns for headless service in k8s: $(service_name).$(k8s_namespace).svc.cluster.local
// ipMaps data like this: { "user-svc":["127.0.0.1:8080","127.0.0.1:8081"] } .
func getDNSForPodIP(svc []*Service) (map[string][]string, error) {
	ipMaps := make(map[string][]string, 10)

	for _, value := range svc {
		dnsForK8sSvc := fmt.Sprintf("%s.%s.svc.cluster.local", value.SvcName, value.Namespace)
		if ipRecords, err := net.LookupIP(dnsForK8sSvc); err != nil {
			return nil, err
		} else {
			for _, ip := range ipRecords {
				ipMaps[value.SvcName] = append(ipMaps[value.SvcName], fmt.Sprintf("%s:%d", ip.String(), value.PodPort))
			}
		}
	}

	return ipMaps, nil
}
