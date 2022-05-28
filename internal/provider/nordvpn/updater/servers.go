// Package nordvpn contains code to obtain the server information
// for the NordVPN provider.
package nordvpn

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/qdm12/gluetun/internal/constants/vpn"
	"github.com/qdm12/gluetun/internal/models"
)

var (
	ErrParseIP          = errors.New("cannot parse IP address")
	ErrNotIPv4          = errors.New("IP address is not IPv4")
	ErrNotEnoughServers = errors.New("not enough servers found")
)

func (u *Updater) GetServers(ctx context.Context, minServers int) (
	servers []models.Server, err error) {
	data, err := fetchAPI(ctx, u.client)
	if err != nil {
		return nil, err
	}

	servers = make([]models.Server, 0, len(data))

	for _, jsonServer := range data {
		if !jsonServer.Features.TCP && !jsonServer.Features.UDP {
			u.warner.Warn("server does not support TCP and UDP for openvpn: " + jsonServer.Name)
			continue
		}

		ip, err := parseIPv4(jsonServer.IPAddress)
		if err != nil {
			return nil, fmt.Errorf("%w for server %s", err, jsonServer.Name)
		}

		number, err := parseServerName(jsonServer.Name)
		if err != nil {
			return nil, err
		}

		server := models.Server{
			VPN:      vpn.OpenVPN,
			Region:   jsonServer.Country,
			Hostname: jsonServer.Domain,
			Number:   number,
			IPs:      []net.IP{ip},
			TCP:      jsonServer.Features.TCP,
			UDP:      jsonServer.Features.UDP,
		}
		servers = append(servers, server)
	}

	if len(servers) < minServers {
		return nil, fmt.Errorf("%w: %d and expected at least %d",
			ErrNotEnoughServers, len(servers), minServers)
	}

	sortServers(servers)

	return servers, nil
}
