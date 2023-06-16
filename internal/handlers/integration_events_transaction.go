package handlers

import (
	"context"
	"database/sql"

	"github.com/v8tix/eda/am"
	"github.com/v8tix/eda/ddd"
	"github.com/v8tix/eda/di"
	"github.com/v8tix/eda/registry"
	orderingpb "github.com/v8tix/mallbots-ordering-proto/pb"
)

func RegisterIntegrationEventHandlersTx(container di.Container) error {
	evtMsgHandler := am.RawMessageHandlerFunc(func(ctx context.Context, msg am.IncomingRawMessage) (err error) {
		ctx = container.Scoped(ctx)
		defer func(tx *sql.Tx) {
			if p := recover(); p != nil {
				_ = tx.Rollback()
				panic(p)
			} else if err != nil {
				_ = tx.Rollback()
			} else {
				err = tx.Commit()
			}
		}(di.Get(ctx, "tx").(*sql.Tx))

		evtHandlers := am.RawMessageHandlerWithMiddleware(
			am.NewEventMessageHandler(
				di.Get(ctx, "registry").(registry.Registry),
				di.Get(ctx, "integrationEventHandlers").(ddd.EventHandler[ddd.Event]),
			),
			di.Get(ctx, "inboxMiddleware").(am.RawMessageHandlerMiddleware),
		)

		return evtHandlers.HandleMessage(ctx, msg)
	})

	subscriber := container.Get("stream").(am.RawMessageStream)

	return subscriber.Subscribe(orderingpb.OrderAggregateChannel, evtMsgHandler, am.MessageFilter{
		orderingpb.OrderCreatedEvent,
	}, am.GroupName("cosec-ordering"))
}
