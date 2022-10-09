package client

import (
	"github.com/pkg/errors"
	"github.com/songgao/water"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	log = logrus.WithField("client", "init")
)

func runCli(cmd *cobra.Command, args []string) error {
	_, err := NewOption().
		WithDefaults().
		WithEnvVariables().
		WithCliFlags(cmd.Flags()).
		Validate()
	if err != nil {
		return errors.Wrap(err, "error when paring flags")
	}

	ifce, err := water.New(water.Config{
		DeviceType: water.TUN,
	})
	if err != nil {
		return errors.Wrap(err, "cannot create TUN device")
	}

	log.Infof("created TUN device: %s", ifce.Name())

	packet := make([]byte, 2000)
	for {
		n, err := ifce.Read(packet)
		if err != nil {
			log.Errorf("cannot read packet: %s", err)
			continue
		}
		log.Printf("received packet: %x\n", packet[:n])
	}

	return nil
}
