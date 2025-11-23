package app

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	config "osom/pkg"
	"osom/pkg/nextbikes"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type LocationAvailability struct {
	LocationName   string
	AvailableBikes int
}

const availabilityRadiusMeters = "500"

func FetchAvailability(ctx context.Context, latitude string, longitude string) ([]LocationAvailability, error) {
	url, _ := url.Parse(config.Config.NextBikesApiUrl)
	q := url.Query()
	q.Set("distance", availabilityRadiusMeters)
	q.Set("lat", latitude)
	q.Set("lng", longitude)
	url.RawQuery = q.Encode()

	client := &http.Client{
		Timeout:   10 * time.Second,
		Transport: otelhttp.NewTransport(http.DefaultTransport),
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var nbResp nextbikes.Response
	if err := json.NewDecoder(res.Body).Decode(&nbResp); err != nil {
		return nil, err
	}

	var out []LocationAvailability
	for _, country := range nbResp.Countries {
		for _, city := range country.Cities {
			for _, p := range city.Places {
				out = append(out, LocationAvailability{
					LocationName:   p.Name,
					AvailableBikes: p.BikesAvailableToRent,
				})
			}
		}
	}

	return out, nil
}
