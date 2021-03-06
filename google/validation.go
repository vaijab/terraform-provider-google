package google

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"net"
	"regexp"
	"strconv"
	"strings"
)

const (
	// Copied from the official Google Cloud auto-generated client.
	ProjectRegex    = "(?:(?:[-a-z0-9]{1,63}\\.)*(?:[a-z](?:[-a-z0-9]{0,61}[a-z0-9])?):)?(?:[0-9]{1,19}|(?:[a-z0-9](?:[-a-z0-9]{0,61}[a-z0-9])?))"
	RegionRegex     = "[a-z](?:[-a-z0-9]{0,61}[a-z0-9])?"
	SubnetworkRegex = "[a-z](?:[-a-z0-9]{0,61}[a-z0-9])?"

	SubnetworkLinkRegex = "projects/(" + ProjectRegex + ")/regions/(" + RegionRegex + ")/subnetworks/(" + SubnetworkRegex + ")$"

	RFC1035NameTemplate = "[a-z](?:[-a-z0-9]{%d,%d}[a-z0-9])"
	CloudIoTIdRegex     = "^[a-zA-Z][-a-zA-Z0-9._+~%]{2,254}$"
)

var (
	// Service account name must have a length between 6 and 30.
	// The first and last characters have different restrictions, than
	// the middle characters. The middle characters length must be between
	// 4 and 28 since the first and last character are excluded.
	ServiceAccountNameRegex = fmt.Sprintf(RFC1035NameTemplate, 4, 28)

	ServiceAccountLinkRegex = "projects/" + ProjectRegex + "/serviceAccounts/" + ServiceAccountNameRegex + "@" + ProjectRegex + "\\.iam\\.gserviceaccount\\.com$"
)

var rfc1918Networks = []string{
	"10.0.0.0/8",
	"172.16.0.0/12",
	"192.168.0.0/16",
}

func validateGCPName(v interface{}, k string) (ws []string, errors []error) {
	re := `^(?:[a-z](?:[-a-z0-9]{0,61}[a-z0-9])?)$`
	return validateRegexp(re)(v, k)
}

func validateRegexp(re string) schema.SchemaValidateFunc {
	return func(v interface{}, k string) (ws []string, errors []error) {
		value := v.(string)
		if !regexp.MustCompile(re).MatchString(value) {
			errors = append(errors, fmt.Errorf(
				"%q (%q) doesn't match regexp %q", k, value, re))
		}

		return
	}
}

func validateRFC1918Network(min, max int) schema.SchemaValidateFunc {
	return func(i interface{}, k string) (s []string, es []error) {

		s, es = validation.CIDRNetwork(min, max)(i, k)
		if len(es) > 0 {
			return
		}

		v, _ := i.(string)
		ip, _, _ := net.ParseCIDR(v)
		for _, c := range rfc1918Networks {
			if _, ipnet, _ := net.ParseCIDR(c); ipnet.Contains(ip) {
				return
			}
		}

		es = append(es, fmt.Errorf("expected %q to be an RFC1918-compliant CIDR, got: %s", k, v))

		return
	}
}

func validateRFC3339Time(v interface{}, k string) (warnings []string, errors []error) {
	time := v.(string)
	if len(time) != 5 || time[2] != ':' {
		errors = append(errors, fmt.Errorf("%q (%q) must be in the format HH:mm (RFC3399)", k, time))
		return
	}
	if hour, err := strconv.ParseUint(time[:2], 10, 0); err != nil || hour > 23 {
		errors = append(errors, fmt.Errorf("%q (%q) does not contain a valid hour (00-23)", k, time))
		return
	}
	if min, err := strconv.ParseUint(time[3:], 10, 0); err != nil || min > 59 {
		errors = append(errors, fmt.Errorf("%q (%q) does not contain a valid minute (00-59)", k, time))
		return
	}
	return
}

func validateRFC1035Name(min, max int) schema.SchemaValidateFunc {
	if min < 2 || max < min {
		return func(i interface{}, k string) (s []string, errors []error) {
			if min < 2 {
				errors = append(errors, fmt.Errorf("min must be at least 2. Got: %d", min))
			}
			if max < min {
				errors = append(errors, fmt.Errorf("max must greater than min. Got [%d, %d]", min, max))
			}
			return
		}
	}

	return validateRegexp(fmt.Sprintf("^"+RFC1035NameTemplate+"$", min-2, max-2))
}

func validateIpCidrRange(v interface{}, k string) (warnings []string, errors []error) {
	_, _, err := net.ParseCIDR(v.(string))
	if err != nil {
		errors = append(errors, fmt.Errorf("%q is not a valid IP CIDR range: %s", k, err))
	}
	return
}

func validateCloudIoTID(v interface{}, k string) (warnings []string, errors []error) {
	value := v.(string)
	if strings.HasPrefix(value, "goog") {
		errors = append(errors, fmt.Errorf(
			"%q (%q) can not start with \"goog\"", k, value))
	}
	if !regexp.MustCompile(CloudIoTIdRegex).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"%q (%q) doesn't match regexp %q", k, value, CloudIoTIdRegex))
	}
	return
}
