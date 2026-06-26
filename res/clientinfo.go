package res

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/transform"
)

type ClientInfo struct {
	ServiceType string
	Connections []Connection
}

type Connection struct {
	Display         string
	Description     string
	Address         string
	Port            int
	Version         int
	LangType        int
	RegistrationWeb string
}

func ParseClientInfoFile(path string) (ClientInfo, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return ClientInfo{}, err
	}
	return ParseClientInfo(data)
}

func ParseClientInfo(data []byte) (ClientInfo, error) {
	data = bytes.TrimPrefix(data, []byte{0xef, 0xbb, 0xbf})

	var raw rawClientInfo
	decoder := xml.NewDecoder(bytes.NewReader(data))
	decoder.CharsetReader = clientInfoCharsetReader
	if err := decoder.Decode(&raw); err != nil {
		return ClientInfo{}, err
	}

	out := ClientInfo{ServiceType: strings.TrimSpace(raw.ServiceType)}
	for _, rawConn := range raw.Connections {
		conn := Connection{
			Display:         strings.TrimSpace(rawConn.Display),
			Description:     strings.TrimSpace(rawConn.Description),
			Address:         strings.TrimSpace(rawConn.Address),
			Port:            parseIntDefault(rawConn.Port, 6900),
			Version:         parseIntDefault(rawConn.Version, 55),
			LangType:        parseIntDefault(rawConn.LangType, 0),
			RegistrationWeb: strings.TrimSpace(rawConn.RegistrationWeb),
		}
		if conn.Address == "" {
			continue
		}
		if conn.Display == "" {
			conn.Display = conn.Address
		}
		out.Connections = append(out.Connections, conn)
	}

	if len(out.Connections) == 0 {
		return out, fmt.Errorf("clientinfo has no usable connection entries")
	}
	return out, nil
}

func clientInfoCharsetReader(charset string, input io.Reader) (io.Reader, error) {
	switch strings.ToLower(strings.TrimSpace(charset)) {
	case "utf-8", "utf8":
		return input, nil
	case "euc-kr", "ks_c_5601-1987", "cp949", "uhc":
		return transform.NewReader(input, korean.EUCKR.NewDecoder()), nil
	default:
		return nil, fmt.Errorf("unsupported clientinfo charset %q", charset)
	}
}

func parseIntDefault(value string, fallback int) int {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	i, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return i
}

type rawClientInfo struct {
	XMLName     xml.Name        `xml:"clientinfo"`
	ServiceType string          `xml:"servicetype"`
	Connections []rawConnection `xml:"connection"`
}

type rawConnection struct {
	Display         string `xml:"display"`
	Description     string `xml:"desc"`
	Address         string `xml:"address"`
	Port            string `xml:"port"`
	Version         string `xml:"version"`
	LangType        string `xml:"langtype"`
	RegistrationWeb string `xml:"registrationweb"`
}
