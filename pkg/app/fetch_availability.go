package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"osom/pkg/app/nextbikes"
	"time"
)

type LocationAvailability struct {
	LocationName   string
	AvailableBikes int
}

func FetchAvailability(latitude string, longitude string) ([]LocationAvailability, error) {
	url, _ := url.Parse("https://api.nextbike.net/maps/nextbike-live.json")
	q := url.Query()
	q.Set("distance", "200")
	q.Set("lat", latitude)
	q.Set("lng", longitude)
	url.RawQuery = q.Encode()

	fmt.Println(url)

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest(http.MethodGet, url.String(), nil)
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
