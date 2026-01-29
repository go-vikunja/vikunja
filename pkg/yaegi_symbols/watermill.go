




package yaegi_symbols

import (
	"context"
	"github.com/ThreeDotsLabs/watermill/message"
	"reflect"
)

func init() {
	Symbols["github.com/ThreeDotsLabs/watermill/message/message"] = map[string]reflect.Value{
		// function, constant and variable definitions
		"ErrOutputInNoPublisherHandler": reflect.ValueOf(&message.ErrOutputInNoPublisherHandler).Elem(),
			"HandlerNameFromCtx": reflect.ValueOf(message.HandlerNameFromCtx),
			"MessageTransformPublisherDecorator": reflect.ValueOf(message.MessageTransformPublisherDecorator),
			"MessageTransformSubscriberDecorator": reflect.ValueOf(message.MessageTransformSubscriberDecorator),
			"NewDefaultRouter": reflect.ValueOf(message.NewDefaultRouter),
			"NewMessage": reflect.ValueOf(message.NewMessage),
			"NewMessageWithContext": reflect.ValueOf(message.NewMessageWithContext),
			"NewRouter": reflect.ValueOf(message.NewRouter),
			"PassthroughHandler": reflect.ValueOf(&message.PassthroughHandler).Elem(),
			"PublishTopicFromCtx": reflect.ValueOf(message.PublishTopicFromCtx),
			"PublisherNameFromCtx": reflect.ValueOf(message.PublisherNameFromCtx),
			"SubscribeTopicFromCtx": reflect.ValueOf(message.SubscribeTopicFromCtx),
			"SubscriberNameFromCtx": reflect.ValueOf(message.SubscriberNameFromCtx),
			
		// type definitions
		"DuplicateHandlerNameError": reflect.ValueOf((*message.DuplicateHandlerNameError)(nil)),
		"Handler": reflect.ValueOf((*message.Handler)(nil)),
		"HandlerFunc": reflect.ValueOf((*message.HandlerFunc)(nil)),
		"HandlerMiddleware": reflect.ValueOf((*message.HandlerMiddleware)(nil)),
		"Message": reflect.ValueOf((*message.Message)(nil)),
		"Messages": reflect.ValueOf((*message.Messages)(nil)),
		"Metadata": reflect.ValueOf((*message.Metadata)(nil)),
		"NoPublishHandlerFunc": reflect.ValueOf((*message.NoPublishHandlerFunc)(nil)),
		"Payload": reflect.ValueOf((*message.Payload)(nil)),
		"Publisher": reflect.ValueOf((*message.Publisher)(nil)),
		"PublisherDecorator": reflect.ValueOf((*message.PublisherDecorator)(nil)),
		"Router": reflect.ValueOf((*message.Router)(nil)),
		"RouterConfig": reflect.ValueOf((*message.RouterConfig)(nil)),
		"RouterPlugin": reflect.ValueOf((*message.RouterPlugin)(nil)),
		"SubscribeInitializer": reflect.ValueOf((*message.SubscribeInitializer)(nil)),
		"Subscriber": reflect.ValueOf((*message.Subscriber)(nil)),
		"SubscriberDecorator": reflect.ValueOf((*message.SubscriberDecorator)(nil)),
		
		// interface wrapper definitions
		"_Publisher": reflect.ValueOf((*_github_com_ThreeDotsLabs_watermill_message_Publisher)(nil)),
		"_SubscribeInitializer": reflect.ValueOf((*_github_com_ThreeDotsLabs_watermill_message_SubscribeInitializer)(nil)),
		"_Subscriber": reflect.ValueOf((*_github_com_ThreeDotsLabs_watermill_message_Subscriber)(nil)),
		
	}
}
// _github_com_ThreeDotsLabs_watermill_message_Publisher is an interface wrapper for Publisher type
	type _github_com_ThreeDotsLabs_watermill_message_Publisher struct {
		IValue interface{}
		WClose func() ( error)
		WPublish func(topic string, messages ...*message.Message) ( error)
		
	}
	func (W _github_com_ThreeDotsLabs_watermill_message_Publisher) Close() ( error) {return W.WClose()
		}
	func (W _github_com_ThreeDotsLabs_watermill_message_Publisher) Publish(topic string, messages ...*message.Message) ( error) {return W.WPublish(topic, messages...)
		}
	
// _github_com_ThreeDotsLabs_watermill_message_SubscribeInitializer is an interface wrapper for SubscribeInitializer type
	type _github_com_ThreeDotsLabs_watermill_message_SubscribeInitializer struct {
		IValue interface{}
		WSubscribeInitialize func(topic string) ( error)
		
	}
	func (W _github_com_ThreeDotsLabs_watermill_message_SubscribeInitializer) SubscribeInitialize(topic string) ( error) {return W.WSubscribeInitialize(topic)
		}
	
// _github_com_ThreeDotsLabs_watermill_message_Subscriber is an interface wrapper for Subscriber type
	type _github_com_ThreeDotsLabs_watermill_message_Subscriber struct {
		IValue interface{}
		WClose func() ( error)
		WSubscribe func(ctx context.Context, topic string) ( <-chan *message.Message,  error)
		
	}
	func (W _github_com_ThreeDotsLabs_watermill_message_Subscriber) Close() ( error) {return W.WClose()
		}
	func (W _github_com_ThreeDotsLabs_watermill_message_Subscriber) Subscribe(ctx context.Context, topic string) ( <-chan *message.Message,  error) {return W.WSubscribe(ctx, topic)
		}
	


