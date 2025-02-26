package pep

import (
	"fmt"
	"net/netip"
	"regexp"
	"strings"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/models"
)

// IP-address is stored in the attributes of the principal.
// See AuthZEN spec Information Model - Subject properties (link is subject to change):
// https://openid.net/specs/authorization-api-1_0-01.html#name-subject-properties
//
// This function is called from the processing of HTTP headers.
// It attempts to find a suitable IP address from the given inputs.
func (c *collector) processForwarded(fwd1, fwd2 string) {
	list1, list2 := strings.Split(fwd1, ","), strings.Split(fwd2, ",")

	if len(list1) < len(list2) {
		if c.processForwardedList(list2) {
			return
		}
		c.processForwardedList(list1)
	} else {
		if c.processForwardedList(list1) {
			return
		}
		c.processForwardedList(list2)
	}
}

func (c *collector) processForwardedList(fwd []string) bool {
	for i := range fwd {
		list := fwdRX.FindStringSubmatch(strings.TrimSpace(fwd[i]))
		for j := 1; j < len(list); j++ {
			if s := list[j]; len(s) > 0 {
				if addr, err := netip.ParseAddr(s); err == nil && validIP(addr) {
					c.parc.Principal.Attributes().AddAttribute(models.AttrClientIP, addr.String())
					return true
				}
			}
		}
	}
	return false
}

func validIP(addr netip.Addr) bool {
	return addr.IsValid() &&
		!addr.IsLoopback() &&
		!addr.IsMulticast() &&
		!addr.IsUnspecified() &&
		!addr.IsInterfaceLocalMulticast() &&
		!addr.IsLinkLocalUnicast() &&
		!addr.IsLinkLocalMulticast()
}

var (
	ip4Match  = `([0-9\.]+)`
	ip6Match  = `([a-f0-9:]+)`
	portMatch = `[:0-9]*`
	fwdMatch1 = fmt.Sprintf(`^(?i:%s|%s)$`, ip4Match, ip6Match)
	fwdMatch2 = fmt.Sprintf(`(?i:(?:for=)(?:%s%s|(?:"\[%s\]%s")))`, ip4Match, portMatch, ip6Match, portMatch)
	fwdRX     = regexp.MustCompile(fmt.Sprintf(`(?:%s|%s)`, fwdMatch1, fwdMatch2))
)
