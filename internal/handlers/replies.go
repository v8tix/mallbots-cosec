package handlers

import (
	"context"

	"github.com/v8tix/eda/am"
	"github.com/v8tix/eda/sec"
	"github.com/v8tix/mallbots-cosec/internal/models"
)

func RegisterReplyHandlers(subscriber am.ReplySubscriber, orchestrator sec.Orchestrator[*models.CreateOrderData]) error {
	replyMsgHandler := am.MessageHandlerFunc[am.IncomingReplyMessage](func(ctx context.Context, replyMsg am.IncomingReplyMessage) error {
		return orchestrator.HandleReply(ctx, replyMsg)
	})
	return subscriber.Subscribe(orchestrator.ReplyTopic(), replyMsgHandler, am.GroupName("cosec-replies"))
}
