package compare

import (
	"net/http"
	"testing"
)

func TestCompare(t *testing.T) {
	cases := map[string]struct {
		want   []byte
		have   []byte
		result bool
		reason string
	}{
		"simple float": {
			want:   []byte(`7.00`),
			have:   []byte(`7.00`),
			result: true,
		},
		"empty object": {
			want:   []byte(`{}`),
			have:   []byte(`{}`),
			result: true,
		},
		"simple int": {
			want:   []byte(`1`),
			have:   []byte(`1`),
			result: true,
		},
		"simple int negative": {
			want:   []byte(`1`),
			have:   []byte(`15`),
			result: false,
			reason: "want: 1, have: 15",
		},
		"slice integer": {
			want:   []byte(`[1, 5, 17]`),
			have:   []byte(`[1, 5, 17]`),
			result: true,
		},
		"slice integer not order": {
			want:   []byte(`[17, 9, 33, 1]`),
			have:   []byte(`[33, 1, 17, 9]`),
			result: true,
		},
		"slice integer diff size": {
			want:   []byte(`[55, 9, 33, 7]`),
			have:   []byte(`[33, 3, 88, 7, 55, 9]`),
			result: true,
		},
		"slice mix integer": {
			want:   []byte(`[1, 5, 9.33]`),
			have:   []byte(`[5, 1, 9.33]`),
			result: true,
		},
		"slice mix type": {
			want:   []byte(`[1, "text", 9.33]`),
			have:   []byte(`["text", 1, 9.33]`),
			result: true,
		},
		"slice text": {
			want:   []byte(`["text", "user name"]`),
			have:   []byte(`["text", "user login", "user name"]`),
			result: true,
		},
		"slice text with utf8": {
			want:   []byte(`["text©", "©", "☮"]`),
			have:   []byte(`["text©", "☮", "©"]`),
			result: true,
		},
		"slice utf8": {
			want:   []byte(`["💦", "💩", "👍"]`),
			have:   []byte(`["👍", "💦", "💩"]`),
			result: true,
		},
		"multiple object": {
			want: []byte(`
			{
				"requestId": "333-333",
				"items": [
					{
						"id": "73.Test00129",
						"externalId": "ext.109484070"
					},
					{
						"id": "125.test.3822273",
						"externalId": "AAP118602397"
					}
				]
			}`),
			have: []byte(`
			{
				"requestId": "333-333",
				"items": [
					{
						"id": "73.Test00128",
						"externalId": "ext.194840740"
					},
					{
						"id": "125.test.3822273",
						"externalId": "AAP118602397"
					},
					{
						"id": "73.Test00129",
						"externalId": "ext.109484070"
					}
				]
			}`),
			result: true,
		},
		"with nil": {
			want:   []byte(`{"key":"value"}`),
			have:   nil,
			result: false,
			reason: "want does not equal nil",
		},
		"with js error have": {
			want:   []byte(`{"key":"value"}`),
			have:   []byte(`{"key":"value"`),
			result: false,
		},
		"with js error want": {
			want:   []byte(`{"key":"value",}`),
			have:   []byte(`{"key":"value"}`),
			result: false,
		},
		"with compare int and string": {
			want:   []byte(`{"key":"value"}`),
			have:   []byte(`{"key":3}`),
			result: false,
			reason: "key 'key' does not match: different types want: string, have: float64",
		},
		"with compare bool and string": {
			want:   []byte(`{"key":true}`),
			have:   []byte(`{"key":"true"}`),
			result: false,
			reason: "key 'key' does not match: different types want: bool, have: string",
		},
		"compare bool type": {
			want:   []byte(`{"bool":true, "string":"string"}`),
			have:   []byte(`{"bool":true, "string":"string"}`),
			result: true,
		},
		"map with slice": {
			want:   []byte(`{"bool":true, "string":"string"}`),
			have:   []byte(`[1, 5]`),
			result: false,
			reason: "different types want: map[string]interface {}, have: []interface {}",
		},
		"different keys": {
			want:   []byte(`{"bool":true, "string":"string"}`),
			have:   []byte(`{"bool":true, "strings":"string"}`),
			result: false,
			reason: "key 'string' not found",
		},
		"slice with map": {
			want:   []byte(`[1,8]`),
			have:   []byte(`{"bool":true, "string":"string"}`),
			result: false,
			reason: "different types want: []interface {}, have: map[string]interface {}",
		},
		"slice count": {
			want:   []byte(`[1,8,7]`),
			have:   []byte(`[1,8]`),
			result: false,
			reason: "different len want: 3, have: 2",
		},
		"slice compare": {
			want:   []byte(`[1,8,7]`),
			have:   []byte(`[1,8,9]`),
			result: false,
		},
		"slice objects success": {
			want: []byte(`
			{
				"id": "5",
				"tags": [
					{
						"id": "1",
						"value": "dog"
					},
					{
						"id": "3",
						"value": "cat"
					}
				]
			}`),
			have: []byte(`
			{
				"id": "5",
				"tags": [
					{
						"id": "2",
						"value": "bird"
					},
					{
						"id": "9",
						"value": "pig"
					},
					{
						"id": "1",
						"value": "dog"
					},

					{
						"id": "6",
						"value": "bug"
					},

					{
						"id": "3",
						"value": "cat"
					},
					{
						"id": "7",
						"value": "mouse"
					}
				]
			}`),
			result: true,
		},
		"slice objects fail": {
			want: []byte(`
			{
				"id": "5",
				"tags": [
					{
						"id": "1",
						"value": "dog"
					},
					{
						"id": "3",
						"value": "cat"
					}
				]
			}`),
			have: []byte(`
			{
				"id": "5",
				"tags": [
					{
						"id": "2",
						"value": "bird"
					},
					{
						"id": "9",
						"value": "pig"
					},
					{
						"id": "1",
						"value": "dogs"
					},

					{
						"id": "6",
						"value": "bug"
					},

					{
						"id": "3",
						"value": "cat"
					},
					{
						"id": "7",
						"value": "mouse"
					}
				]
			}`),
			result: false,
			reason: "key 'tags' does not match: value not found in slice: map[id:1 value:dog]",
		},
		"nested empy slice": {
			want: []byte(`
			{
				"id": "5",
				"tags": []
			}`),
			have: []byte(`
			{
				"id": "5",
				"tags": ["dog"]
			}`),
			result: true,
		},
		"nil compare": {
			want: []byte(`
			{
				"id": "5",
				"tags": null
			}`),
			have: []byte(`
			{
				"id": "5",
				"tags": null
			}`),
			result: true,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			cr, err := JSONCompare(tc.want, tc.have)
			if cr != tc.result {
				t.Fatalf("not equal:\nwant: %s\nhave: %s\n", tc.want, tc.have)
			}
			if len(tc.reason) == 0 {
				return
			}

			if err == nil {
				t.Fatalf("empty error")
			}

			if err.Error() != tc.reason {
				t.Fatalf("not equal:\nwant: %s\nhave: %s\n", tc.reason, err.Error())
			}
		})
	}

	res, _ := compareValues(http.Header{}, nil)
	if res {
		t.Fatalf("not equal:\nwant: %v\nhave: %v\n", false, res)
	}
}
