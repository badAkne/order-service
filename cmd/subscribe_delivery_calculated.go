package cmd

import (
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/badAkne/order-service/internal/app/builder"
)

const (
	cmdSubscribeDeliveryCalculatedUsage = "Starts the consumer for calculating delivery price"

	cmdSubscribeDeliveryCalculatedDescription = `
Initializes and starts consumer, that listens specified topic
for incoming Kafka dewlivery events.
`
)

func DeliveryConsumer() *cli.Command {
	return &cli.Command{
		Name:            "DeliveryConsumer",
		Aliases:         []string{"delivery"},
		Usage:           cmdSubscribeDeliveryCalculatedUsage,
		Description:     strings.TrimSpace(cmdSubscribeDeliveryCalculatedDescription),
		Action:          cmdDeliveryConsumer,
		HideHelpCommand: true,
	}
}

func cmdDeliveryConsumer(cCtx *cli.Context) error {
	app := builder.NewBuilder(cCtx)
	app.BuildConfig()

	app.BuildRepoConnPostgres()
	app.BuildBrokerKafka()

	app.BuildCatalogClient()
	app.BuildRepoOrder()

	app.BuildServiceOrder()

	app.BuildHandlerDeliveryCalculated()

	app.BuildProcEventDeliveryCalculated()

	app.Run()

	return nil
}
