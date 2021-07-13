---
date: 2018-10-13T19:26:34+02:00
title: "Events and Listeners"
draft: false
menu:
  sidebar:
    parent: "development"
---

# Events and Listeners

Vikunja provides a simple observer pattern mechanism through events and listeners.
The basic principle of events is always the same: Something happens (=An event is fired) and something reacts to it (=A listener is called).

Vikunja supports this principle through the `events` package.
It is built upon the excellent [watermill](https://watermill.io) library.

Currently, it only supports dispatching events through Go Channels which makes it configuration-less.
More methods of dispatching events (like kafka or rabbitmq) are available in watermill and could be enabled with a PR.

This document explains how events and listeners work in Vikunja, how to use them and how to create new ones.

{{< table_of_contents >}}

## Events

### Definition

Each event has to implement this interface:

{{< highlight golang >}}
type Event interface {
    Name() string
}
{{< /highlight >}}

An event can contain whatever data you need.

When an event is dispatched, all of the data it contains will be marshaled into json for dispatching.
You then get the event with all its data back in the listener, see below.

#### Naming Convention

Event names should roughly have the entity they're dealing with on the left and the action on the right of the name, separated by `.`.
There's no limit to how "deep" or specifig an event name can be.

The name should have the most general concept it's describing at the left, getting more specific on the right of it.

#### Location

All events for a package should be declared in the `events.go` file of that package.

### Creating a New Event

The easiest way to create a new event is to generate it with mage:

```
mage dev:make-event <event-name> <package>
```

The function takes the name of the event as the first argument and the package where the event should be created as the second argument.
Events will be appended to the `pkg/<module>/events.go` file.
Both parameters are mandatory.

The event type name is automatically camel-cased and gets the `Event` suffix if the provided name does not already have one.
The event name is derived from the type name and stripped of the `.event` suffix.

The generated event will look something like the example below.

### Dispatching events

To dispatch an event, simply call the `events.Dispatch` method and pass in the event as parameter.

### Example

The `TaskCreatedEvent` is declared in the `pkg/models/events.go` file as follows:

{{< highlight golang >}}
// TaskCreatedEvent represents an event where a task has been created
type TaskCreatedEvent struct {
    Task *Task
    Doer web.Auth
}

// Name defines the name for TaskCreatedEvent
func (t *TaskCreatedEvent) Name() string {
    return "task.created"
}
{{< /highlight >}}

It is dispatched in the `createTask` function of the `models` package:

{{< highlight golang >}}
func createTask(s *xorm.Session, t *Task, a web.Auth, updateAssignees bool) (err error) {

    // ...
    
    err = events.Dispatch(&TaskCreatedEvent{
        Task: t,
        Doer: a,
    })
    
    // ...
}
{{< /highlight >}}

As you can see, the curent task and doer are injected into it.

### Special Events

#### `BootedEvent`

Once Vikunja is fully initialized, right before the api web server is started, this event is fired.

## Listeners

A listener is a piece of code that gets executed asynchronously when an event is dispatched.

A single event can have multiple listeners who are independent of each other.

### Definition

All listeners must implement this interface:

{{< highlight golang >}}
// Listener represents something that listens to events
type Listener interface {
    Handle(msg *message.Message) error
    Name() string
}
{{< /highlight >}}

The `Handle` method is executed when the event this listener listens on is dispatched. 
* As the single parameter, it gets the payload of the event, which is the event struct when it was dispatched decoded as json object and passed as a slice of bytes.
To use it you'll need to unmarshal it. Unfortunately there's no way to pass an already populated event object to the function because we would not know what type it has when parsing it.
* If the handler returns an error, the listener is retried 5 times, with an exponentional back-off period in between retries.
If it still fails after the fifth retry, the event is nack'd and it's up to the event dispatcher to resend it.
You can learn more about this mechanism in the [watermill documentation](https://watermill.io/docs/middlewares/#retry).

The `Name` method needs to return a unique listener name for this listener.
It should follow the same convention as event names, see above.

### Creating a New Listener

The easiest way to create a new listener for an event is with mage:

```
mage dev:make-listener <listener-name> <event-name> <package>
```

This will create a new listener type in the `pkg/<package>/listners.go` file and implement the `Handle` and `Name` methods.
It will also pre-generate some boilerplate code to unmarshal the event from the payload.

Furthermore, it will register the listener for its event in the `RegisterListeners()` method of the same file.
This function is called at startup and has to contain all events you want to listen for.

### Listening for Events

To listen for an event, you need to register the listener for the event it should be called for.
This usually happens in the `RegisterListeners()` method in `pkg/<package>/listners.go` which is called at start up.

The listener will never be executed if it hasn't been registered.

See the example below.

### Example

{{< highlight golang >}}
// RegisterListeners registers all event listeners
func RegisterListeners() {
    events.RegisterListener((&ListCreatedEvent{}).Name(), &IncreaseListCounter{})
}

// IncreaseTaskCounter represents a listener
type IncreaseTaskCounter struct {}

// Name defines the name for the IncreaseTaskCounter listener
func (s *IncreaseTaskCounter) Name() string {
    return "task.counter.increase"
}

// Hanlde is executed when the event IncreaseTaskCounter listens on is fired
func (s *IncreaseTaskCounter) Handle(payload message.Payload) (err error) {
    return keyvalue.IncrBy(metrics.TaskCountKey, 1)
}
{{< /highlight >}}

## Testing

When testing, you should call the `events.Fake()` method in the `TestMain` function of the package you want to test.
This prevents any events from being fired and lets you assert an event has been dispatched like so:

{{< highlight golang >}}
events.AssertDispatched(t, &TaskCreatedEvent{})
{{< /highlight >}}
