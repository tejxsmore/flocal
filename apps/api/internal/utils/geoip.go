package utils

import (
	"net"

	"github.com/oschwald/geoip2-golang"
)

type GeoIP struct {
	reader *geoip2.Reader
}

func NewGeoIP(dbPath string) (*GeoIP, error) {
	reader, err := geoip2.Open(dbPath)
	if err != nil {
		return nil, err
	}

	return &GeoIP{reader: reader}, nil
}

func (g *GeoIP) Close() error {
	return g.reader.Close()
}

func (g *GeoIP) CountryCode(ip net.IP) (string, bool) {
	if ip == nil {
		return "", false
	}

	record, err := g.reader.Country(ip)
	if err != nil {
		return "", false
	}

	if record.Country.IsoCode == "" {
		return "", false
	}

	return record.Country.IsoCode, true
}
