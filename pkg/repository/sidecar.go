package repository

import (
	"context"
	"fmt"

	"github.com/rueian/rueidis"
)

func GetServicesNextHop(ctx context.Context, client rueidis.Client) (map[string]string, error) {
	// scan all keys
	scanRes, err := client.Do(ctx, client.B().Scan().Cursor(0).Build()).ToAny()
	if err != nil {
		return nil, err
	}
	allKeys, ok := scanRes.([]interface{})[1].([]interface{})
	if !ok {
		return nil, fmt.Errorf("parsing all keys error")
	}
	// get all service next hop
	nextHops := make(map[string]string)
	for _, val := range allKeys {
		service := val.(string)
		nextHop, err := client.Do(ctx, client.B().Get().Key(service).Build()).ToString()
		if err != nil {
			return nil, fmt.Errorf("get service %s next hop error! ", service)
		}
		nextHops[service] = nextHop
	}
	return nextHops, nil
}
