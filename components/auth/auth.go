package auth

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"time"

	"github.com/bupt-narc/rinp/pkg/util/iplist"
	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v5"
	"github.com/pkg/errors"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/cmd"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/models/schema"
	"github.com/rueian/rueidis"
)

var (
	UserIPCIDR        *net.IPNet
	ServerCIDR        []string
	FirstProxyAddress string
	SchedulerAddress  string
	redisConn         rueidis.Client
)

// CLI flags
var (
	redisAddr = "redis:6379"
)

// TODO: make JWT expiry time shorter

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

	fmt.Printf("auth token valid for %d\n", app.Settings().UserAuthToken.Duration)

	fmt.Println(redisAddr)
	redisConn, err = rueidis.NewClient(rueidis.ClientOption{
		InitAddress: []string{redisAddr},
		SelectDB:    0,
	})
	if err != nil {
		return err
	}
	defer redisConn.Close()

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

	app.OnUserAuthRequest().Add(func(e *core.UserAuthEvent) error {
		ctx := context.Background()

		vip := e.User.Profile.GetDataValue("vip").(string)

		token, err := jwt.Parse(e.Token, func(token *jwt.Token) (interface{}, error) {
			return []byte(e.User.TokenKey + app.Settings().UserAuthToken.Secret), nil
		})
		if err != nil {
			return errors.Wrapf(err, "error when parsing jwt")
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			expireSeconds := int64(claims["exp"].(float64)) - time.Now().Unix()
			err := redisConn.Do(ctx, redisConn.B().Set().Key(vip).Value(iplist.ToString(FirstProxyAddress)).ExSeconds(expireSeconds).Build()).Error()
			if err != nil {
				return err
			}
		}

		return nil
	})

	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		// create vip filed for user profile
		{
			collection, err := app.Dao().FindCollectionByNameOrId("systemprofiles0")
			if err != nil {
				return errors.Wrapf(err, "cannot find systemprofiles0")
			}

			vipField := collection.Schema.GetFieldByName("vip")
			if vipField == nil {
				collection.Schema.AddField(&schema.SchemaField{
					System:   true,
					Id:       RandomString(8),
					Name:     "vip",
					Type:     schema.FieldTypeText,
					Required: false,
					Unique:   false,
					Options:  nil,
				})

				err = app.Dao().SaveCollection(collection)
				if err != nil {
					return errors.Wrapf(err, "cannot save collection")
				}
			}
		}

		// add route for server CIDR, first proxy ip, scheduler ip
		{
			_, err := e.Router.AddRoute(echo.Route{
				Method: http.MethodGet,
				Path:   "/api/v1/rinp",
				Handler: func(c echo.Context) error {
					return c.JSON(http.StatusOK, map[string]interface{}{
						"server_cidr":         ServerCIDR,
						"first_proxy_address": FirstProxyAddress,
						"scheduler_address":   SchedulerAddress,
					})
				},
				Middlewares: []echo.MiddlewareFunc{apis.RequireAdminOrUserAuth()},
			})
			if err != nil {
				return errors.Wrapf(err, "cannot add route")
			}
		}
		return nil
	})

	app.RootCmd.AddCommand(cmd.NewServeCommand(app, false))
	app.RootCmd.AddCommand(cmd.NewMigrateCommand(app))

	flags := app.RootCmd.PersistentFlags()

	// FIXME: unfunctional flag
	flags.StringVarP(&redisAddr, "redis", "r", "127.0.0.1:6379", "Redis address")

	err = app.Execute()

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
