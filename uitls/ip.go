package uitls

import (
	"fmt"
	"net"
	"net/url"
	"strings"
)

//GetLocalIP get host ip by access external server
func GetLocalIP(extip string) string {
	conn, err := net.Dial("tcp", extip)
	if err != nil {
		fmt.Println(err.Error())
		return ""
	}
	defer conn.Close()

	localip := conn.LocalAddr().String()
	ipParts := strings.Split(localip, ":")
	fmt.Println("localip=", localip)
	return ipParts[0]
}

//GetLocalIPByAccessHTTPServer http://extip:port/path user extip:path to get localip
func GetLocalIPByAccessHTTPServer(server string) string {
	u, err := url.Parse(server)
	if err != nil {
		return ""
	}
	return GetLocalIP(u.Host)
}
