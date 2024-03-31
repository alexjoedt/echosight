package echosight

import (
	"github.com/alexjoedt/echosight/internal/validator"
	"github.com/google/uuid"
)

type Validator interface {
	Validate() bool
}

func ValidateTokenPlainText(v *validator.Validator, tokenPlaintext string) {
	v.Check(tokenPlaintext != "", "token", "must be provided")
	v.Check(len(tokenPlaintext) == 26, "token", "must be 26 bytes long")
}

func ValidateEmail(v *validator.Validator, email string) {
	v.Check(email != "", "email", "must be provided")
	v.Check(validator.Matches(email, validator.RegExEmail), "email", "invalid email address")
}

func ValidateDetector(v *validator.Validator, detector *Detector) {
	v.Check(len(detector.Name) > 3, "name", "name too short")
	v.Check(uuid.Validate(detector.HostID.String()) == nil, "hostID", "invalid host ID")
	v.Check(ValidateDetectorConfig(detector), "config", "invalid config for type "+detector.Type.String())
	v.Check(ValidateDetectorType(detector), "type", "unsupported detector type")
}

func ValidateHost(v *validator.Validator, host *Host) {
	v.Check(len(host.Name) > 3, "name", "name too short")

	v.Check(validAddressType(host.AddressType), "AddressType", "invalid address type")

	if host.Address != "" {
		v.Check(validator.IsIP(host.Address), "Address", "invalid address, must be IPv4 or IPv6")
	}
}

func validAddressType(addressType AddressType) bool {
	return addressType == AddressTypeIPv4 || addressType == AddressTypeIPv6
}
