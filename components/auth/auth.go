package auth

import (
	"context"
	"fmt"
	"github.com/bupt-narc/rinp/pkg/util/iplist"
	"math/rand"
	"net"
	"os"
	"time"

	"github.com/pkg/errors"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/cmd"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/models/schema"
	"github.com/rueian/rueidis"
	"github.com/sirupsen/logrus"
)

var (
	UserIPCIDR        *net.IPNet
	ServerCIDR        []string
	FirstProxyAddress string
	SchedulerAddress  string
	redisClient       rueidis.Client
)

// CLI flags
var (
	redisAddr = "redis:6379"
)

func init() {
	// init seed
	rand.Seed(time.Now().UTC().UnixNano())

	var (
		err error
	)
	// United States Department of Defense Network Information Center
	_, UserIPCIDR, err = net.ParseCIDR("7.0.0.0/8")
	if err != nil {
		panic(err)
	}
	ServerCIDR = []string{"11.22.33.44/24"}
	FirstProxyAddress = "proxy1:5114"
	SchedulerAddress = "11.22.33.55:5525"
}

func Execute() error {
	var (
		err error
	)

	app := pocketbase.New()

	addFlags(app)

	fmt.Println(redisAddr)
	redisClient, err = rueidis.NewClient(rueidis.ClientOption{
		InitAddress: []string{redisAddr},
		SelectDB:    0,
	})
	if err != nil {
		return err
	}
	defer redisClient.Close()

	// create vip filed for collection users
	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {

		collection, err := app.Dao().FindCollectionByNameOrId("users")
		if err != nil {
			return errors.Wrapf(err, "cannot find users")
		}

		vipField := collection.Schema.GetFieldByName("vip")
		if vipField == nil {
			collection.Schema.AddField(&schema.SchemaField{
				System:   true,
				Id:       RandomString(8),
				Name:     "vip",
				Type:     schema.FieldTypeText,
				Required: false,
				Unique:   true,
				Options:  nil,
			})
			collection.Schema.AddField(&schema.SchemaField{
				System:   true,
				Id:       RandomString(8),
				Name:     "serverCIDR",
				Type:     schema.FieldTypeText,
				Required: false,
				Options:  nil,
			})
			collection.Schema.AddField(&schema.SchemaField{
				System:   true,
				Id:       RandomString(8),
				Name:     "firstProxyAddress",
				Type:     schema.FieldTypeText,
				Required: false,
				Options:  nil,
			})
			collection.Schema.AddField(&schema.SchemaField{
				System:   true,
				Id:       RandomString(8),
				Name:     "schedulerAddress",
				Type:     schema.FieldTypeText,
				Required: false,
				Options:  nil,
			})
			err = app.Dao().SaveCollection(collection)
			if err != nil {
				return errors.Wrapf(err, "cannot save collection")
			}
		}

		return nil
	})

	app.OnRecordBeforeCreateRequest("users").Add(func(e *core.RecordCreateEvent) error {
		e.Record.Set("serverCIDR", ServerCIDR)
		e.Record.Set("firstProxyAddress", FirstProxyAddress)
		e.Record.Set("schedulerAddress", SchedulerAddress)
		ip, err := UniqueRandomIP(app, UserIPCIDR)
		if err != nil {
			return err
		}
		e.Record.Set("vip", ip)
		return nil
	})

	app.OnRecordAfterAuthWithPasswordRequest().Add(func(e *core.RecordAuthWithPasswordEvent) error {
		if e.Collection.Name == "users" {
			host := e.Record.Get("vip").(string)
			ctx, _ := context.WithCancel(context.Background())
			redisClient.Do(ctx, redisClient.B().Set().Key(host).Value(iplist.ToString(FirstProxyAddress)).Build()).Error()
		}
		return nil
	})

	app.RootCmd.AddCommand(cmd.NewServeCommand(app, false))

	err = app.Execute()

	return err
}

func addFlags(app *pocketbase.PocketBase) {
	flags := app.RootCmd.PersistentFlags()
	flags.StringVarP(&redisAddr, "redis", "r", redisAddr, "Redis address")
	err := app.RootCmd.ParseFlags(os.Args[1:])
	if err != nil {
		logrus.Errorf("error when parsing flags: %s", err)
	}
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

	randomIP := ip + (rand.Uint32() & mask)

	for i := 0; i < 4; i++ {
		ipv4[i] = byte(randomIP >> (8 * (3 - i)))
	}

	return ipv4
}

func UniqueRandomIP(app *pocketbase.PocketBase, cidr *net.IPNet) (net.IP, error) {

	records, err := app.Dao().FindRecordsByExpr("users", dbx.Like("vip", ""))
	if err != nil {
		return nil, err
	}
	ipMap := make(map[string]bool)
	for _, r := range records {
		ipStr := r.Get("vip").(string)
		ipMap[ipStr] = true
	}

	var ip net.IP
	var i int
	// Loop until a unique IP is found
	for i = 0; i < 10000; i++ {
		ip = RandomIP(cidr)
		// ip should not end with 0 or 255
		if ip[3] == 0 || ip[3] == 255 {
			continue
		}
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
