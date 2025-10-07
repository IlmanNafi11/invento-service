package helper

import (
	"errors"
	"strings"
)

type EmailDomainInfo struct {
	IsValid    bool
	Subdomain  string
	RoleName   string
}

func ValidatePolijeEmail(email string) (*EmailDomainInfo, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return nil, errors.New("format email tidak valid")
	}

	domain := parts[1]
	
	if !strings.HasSuffix(domain, "polije.ac.id") {
		return nil, errors.New("hanya email dengan domain polije.ac.id yang dapat digunakan")
	}

	subdomain := ""
	if strings.Contains(domain, ".") {
		domainParts := strings.Split(domain, ".")
		if len(domainParts) >= 3 {
			subdomain = domainParts[0]
		}
	}

	info := &EmailDomainInfo{
		IsValid:   true,
		Subdomain: subdomain,
	}

	switch subdomain {
	case "student":
		info.RoleName = "mahasiswa"
	case "teacher":
		info.RoleName = "dosen"
	default:
		return nil, errors.New("subdomain email tidak valid, gunakan student atau teacher")
	}

	return info, nil
}
