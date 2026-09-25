// pkgcommon's init() exits the process when the working directory has no .env
// file; run these with one present (an empty file is enough).
package pkgcommon

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"
)

func TestInAppNotificationBuild(t *testing.T) {
	expires := time.Date(2026, 9, 25, 21, 0, 0, 0, time.UTC)
	n := InAppNotification{
		Topic:            "in-app-notification",
		Text:             "Zain sent you a friend request",
		Type:             "friend_request",
		TypeID:           "fr-1",
		Recipients:       []string{"u-2"},
		Data:             json.RawMessage(`{"id":"fr-1","sender":{"id":"u-1"}}`),
		ActorID:          "u-1",
		CollapseKey:      "friend_request:u-1:u-2",
		Persona:          "member",
		ActionType:       "friend_request",
		ActionDescriptor: json.RawMessage(`{"actions":[]}`),
		ActionExpiresAt:  &expires,
	}

	input, err := n.build()
	if err != nil {
		t.Fatalf("build: %v", err)
	}

	// Only routing metadata rides as attributes, however many fields the document has.
	var names []string
	for k := range input.MessageAttributes {
		names = append(names, k)
	}
	if len(names) != 2 || input.MessageAttributes["schema"] == nil || input.MessageAttributes["type"] == nil {
		t.Fatalf("attributes = %v, want exactly schema and type", names)
	}
	if got := *input.MessageAttributes["schema"].StringValue; got != InAppNotificationSchema {
		t.Fatalf("schema = %q", got)
	}

	// The body round-trips into the same document the consumer decodes.
	var got InAppNotification
	if err := json.Unmarshal([]byte(*input.Message), &got); err != nil {
		t.Fatalf("body is not the document: %v", err)
	}
	got.Topic = n.Topic
	if !reflect.DeepEqual(got, n) {
		t.Fatalf("round trip mismatch:\n got %+v\nwant %+v", got, n)
	}
}

func TestInAppNotificationBuildEvent(t *testing.T) {
	n := InAppNotification{Topic: "in-app-notification", Type: "friend_request", TypeID: "fr-1", Event: "action_status_changed", ActionStatus: "accept"}
	input, err := n.build()
	if err != nil {
		t.Fatalf("an event message needs no text or recipients: %v", err)
	}
	if input.MessageAttributes["event"] == nil || *input.MessageAttributes["event"].StringValue != "action_status_changed" {
		t.Fatal("event must be a routing attribute")
	}
	var doc map[string]json.RawMessage
	json.Unmarshal([]byte(*input.Message), &doc)
	if string(doc["recipients"]) != "[]" {
		t.Fatalf(`recipients = %s, want [] (never null)`, doc["recipients"])
	}
}

func TestInAppNotificationBuildRejects(t *testing.T) {
	valid := InAppNotification{Topic: "t", Text: "x", Type: "a", TypeID: "b", Recipients: []string{"u"}}
	cases := map[string]func(n *InAppNotification){
		"no topic":            func(n *InAppNotification) { n.Topic = "" },
		"no type":             func(n *InAppNotification) { n.Type = "" },
		"no type_id":          func(n *InAppNotification) { n.TypeID = "" },
		"no text":             func(n *InAppNotification) { n.Text = "" },
		"bad data":            func(n *InAppNotification) { n.Data = json.RawMessage(`{`) },
		"bad descriptor":      func(n *InAppNotification) { n.ActionDescriptor = json.RawMessage(`[`) },
		"over the size limit": func(n *InAppNotification) { n.Text = string(make([]byte, maxSNSMessageSize)) },
	}
	for name, mutate := range cases {
		n := valid
		mutate(&n)
		if _, err := n.build(); err == nil {
			t.Errorf("%s: build succeeded, want an error", name)
		}
	}
}
