package auth

import (
	"fmt"
	"math/rand"
	"net"
	"time"

	"github.com/pkg/errors"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/models/schema"
)

var (
	UserIPCIDR *net.IPNet
)

func init() {
	var (
		err error
	)
	// United States Department of Defense Network Information Center
	_, UserIPCIDR, err = net.ParseCIDR("7.0.0.0/8")
	if err != nil {
		panic(err)
	}
}

func Execute() error {
	app := pocketbase.New()

	app.OnUserAfterCreateRequest().Add(func(e *core.UserCreateEvent) error {
		userID := e.User.Id
		user, err := app.Dao().FindUserById(userID)
		if err != nil {
			return errors.Wrapf(err, "cannot find user to assign vip")
		}

		ip, err := UniqueRandomIP(app, UserIPCIDR)
		if err != nil {
			return err
		}

		records, err := app.Dao().FindUserRelatedRecords(user)
		if err != nil {
			return errors.Wrapf(err, "cannot find records")
		}
		for _, r := range records {
			r.SetDataValue("vip", ip.String())
			err := app.Dao().SaveRecord(r)
			if err != nil {
				return errors.Wrapf(err, "cannot save record %s", r.Id)
			}
		}
		return nil
	})

	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		collection, err := app.Dao().FindCollectionByNameOrId("systemprofiles0")
		if err != nil {
			return errors.Wrapf(err, "cannot find systemprofiles0")
		}

		oldField := collection.Schema.GetFieldByName("vip")
		if oldField != nil {
			return nil
		}

		collection.Schema.AddField(&schema.SchemaField{
			System:   true,
			Id:       RandomString(8),
			Name:     "vip",
			Type:     schema.FieldTypeText,
			Required: false,
			Unique:   false,
			Options:  nil,
		})

		// TODO: add route for server CIDR, first proxy ip, controller ip

		return app.Dao().SaveCollection(collection)
	})

	err := app.Start()

	return err
}

// Generate random alpha-numeric string
func RandomString(length int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	b := make([]rune, length)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// Generate random IP address in CIDR notation
func RandomIP(cidr *net.IPNet) net.IP {
	var ip uint32
	ipv4 := net.ParseIP(cidr.IP.String()).To4()
	// To uint32
	for i := 0; i < 4; i++ {
		ip += uint32(ipv4[i]) << (8 * (3 - i))
	}

	ones, size := cidr.Mask.Size()
	mask := uint32(0xFFFFFFFF<<(size-ones)) ^ 0xFFFFFFFF

	// init seed
	rand.Seed(time.Now().UTC().UnixNano())
	randomIP := ip + (rand.Uint32() & mask)

	for i := 0; i < 4; i++ {
		ipv4[i] = byte(randomIP >> (8 * (3 - i)))
	}

	return ipv4
}

func UniqueRandomIP(app *pocketbase.PocketBase, cidr *net.IPNet) (net.IP, error) {
	collection, err := app.Dao().FindCollectionByNameOrId("systemprofiles0")
	if err != nil {
		return nil, errors.Wrapf(err, "cannot find systemprofiles0")
	}
	//field := collection.Schema.GetFieldByName("vip")

	records, err := app.Dao().FindRecordsByExpr(collection, dbx.Like("vip", ""))
	if err != nil {
		return nil, err
	}
	ipMap := make(map[string]bool)
	for _, r := range records {
		ipStr := r.Data()["vip"].(string)
		ipMap[ipStr] = true
	}

	var ip net.IP
	var i int
	// Loop until a unique IP is found
	for i = 0; i < 10000; i++ {
		ip = RandomIP(cidr)
		_, ok := ipMap[ip.String()]
		if !ok {
			break
		}
	}
	if i >= 10000 {
		return nil, fmt.Errorf("retry uplimit exceeded")
	}

	return ip, nil
}
