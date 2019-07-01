// Copyright (c) 2015 Andy Leap, Google
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS
// IN THE SOFTWARE.

package microformats

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func Test_Datetime_SetDate(t *testing.T) {
	var d datetime

	// ref is the first time.Time value set.  All subsequent values set
	// should have no effect.
	ref := time.Date(2000, 1, 2, 0, 0, 0, 0, time.UTC)

	for _, tt := range []time.Time{
		ref,
		time.Date(2003, 4, 5, 0, 0, 0, 0, time.UTC),
		time.Date(2003, 4, 5, 6, 7, 8, 0, time.UTC),
		{},
	} {
		d.setDate(tt.Year(), tt.Month(), tt.Day())

		// check that has* flags are set properly
		if got, want := d.hasDate, true; got != want {
			t.Errorf("datetime(%v) hasDate: %t, want %t", tt, got, want)
		}
		if got, want := d.hasTime, false; got != want {
			t.Errorf("datetime(%v) hasTime: %t, want %t", tt, got, want)
		}
		if got, want := d.hasTZ, false; got != want {
			t.Errorf("datetime(%v) hasTZ: %t, want %t", tt, got, want)
		}

		if got, want := d.t, ref; !got.Equal(want) {
			t.Errorf("datetime(%v) has time: %v, want %v", tt, got, want)
		}
	}
}

func Test_Datetime_SetTime(t *testing.T) {
	var d datetime

	// ref is the first time.Time value set.  All subsequent values set
	// should have no effect.  date set to 0001-01-01 to match zero value.
	ref := time.Date(1, 1, 1, 1, 2, 3, 0, time.UTC)

	for _, tt := range []time.Time{
		ref,
		time.Date(0, 0, 0, 4, 5, 6, 0, time.UTC),
		time.Date(2000, 1, 3, 4, 5, 6, 0, time.UTC),
		{},
	} {
		d.setTime(tt.Hour(), tt.Minute(), tt.Second())

		// check that has* flags are set properly
		if got, want := d.hasDate, false; got != want {
			t.Errorf("datetime(%v) hasDate: %t, want %t", tt, got, want)
		}
		if got, want := d.hasTime, true; got != want {
			t.Errorf("datetime(%v) hasTime: %t, want %t", tt, got, want)
		}
		if got, want := d.hasTZ, false; got != want {
			t.Errorf("datetime(%v) hasTZ: %t, want %t", tt, got, want)
		}

		if got, want := d.t, ref; !got.Equal(want) {
			t.Errorf("datetime(%v) has time: %v, want %v", tt, got, want)
		}
	}
}

func Test_Datetime_SetTZ(t *testing.T) {
	var d datetime

	// ref is the first time.Time value set.  All subsequent values set
	// should have no effect.  date set to 0001-01-01 to match zero value.
	ref := time.Date(1, 1, 1, 0, 0, 0, 0, time.FixedZone("", -8*60*60))

	for _, tt := range []time.Time{
		ref,
		time.Date(2000, 1, 3, 4, 5, 6, 0, time.UTC),
		{},
	} {
		d.setTZ(tt.Location())

		// check that has* flags are set properly
		if got, want := d.hasDate, false; got != want {
			t.Errorf("datetime(%v) hasDate: %t, want %t", tt, got, want)
		}
		if got, want := d.hasTime, false; got != want {
			t.Errorf("datetime(%v) hasTime: %t, want %t", tt, got, want)
		}
		if got, want := d.hasTZ, true; got != want {
			t.Errorf("datetime(%v) hasTZ: %t, want %t", tt, got, want)
		}

		if got, want := d.t, ref; !got.Equal(want) {
			t.Errorf("datetime(%v) has time: %v, want %v", tt, got, want)
		}
	}
}

func Test_Datetime_String(t *testing.T) {
	ref := time.Date(2006, 1, 2, 15, 04, 05, 0, time.UTC)

	tests := []struct {
		datetime datetime
		want     string
	}{
		{datetime{t: ref}, ""},
		{datetime{t: ref, hasDate: true}, "2006-01-02"},
		{datetime{t: ref, hasDate: true, hasTime: true}, "2006-01-02 15:04"},
		{
			datetime{t: ref, hasDate: true, hasTime: true, hasSeconds: true},
			"2006-01-02 15:04:05",
		},
		{
			datetime{t: ref, hasDate: true, hasTime: true, hasTZ: true},
			"2006-01-02 15:04Z",
		},
		{
			datetime{t: ref, hasDate: true, hasTime: true, hasSeconds: true, hasTZ: true},
			"2006-01-02 15:04:05Z",
		},
		{
			datetime{
				t:       time.Date(2006, 1, 2, 15, 04, 05, 0, time.FixedZone("", -5*60*60)),
				hasDate: true, hasTime: true, hasSeconds: true, hasTZ: true,
			},
			"2006-01-02 15:04:05-0500",
		},
	}

	for _, tt := range tests {
		if got, want := tt.datetime.String(), tt.want; got != want {
			t.Errorf("datetime(%v).String returned %v, want %v", tt.datetime, got, want)
		}
	}
}

func Test_Datetime_Parse(t *testing.T) {
	tests := []struct {
		input string
		want  time.Time
	}{
		{"2000-01-02T03:04:05Z", time.Date(2000, 1, 2, 3, 4, 5, 0, time.UTC)},
		{"2000-01-02t03:04Z", time.Date(2000, 1, 2, 3, 4, 0, 0, time.UTC)},

		{"2000-01-02", time.Date(2000, 1, 2, 0, 0, 0, 0, time.UTC)},
		{"2000-145", time.Date(2000, 5, 24, 0, 0, 0, 0, time.UTC)},

		{"03:04:05-05:00", time.Date(1, 1, 1, 3, 4, 5, 0, time.FixedZone("", -5*60*60))},
		{"03:04:05+05:00", time.Date(1, 1, 1, 3, 4, 5, 0, time.FixedZone("", 5*60*60))},
		{"03:04:05-0500", time.Date(1, 1, 1, 3, 4, 5, 0, time.FixedZone("", -5*60*60))},
		{"03:04:05+0500", time.Date(1, 1, 1, 3, 4, 5, 0, time.FixedZone("", 5*60*60))},
		{"03:04:05Z", time.Date(1, 1, 1, 3, 4, 5, 0, time.UTC)},
		{"03:04:05", time.Date(1, 1, 1, 3, 4, 5, 0, time.UTC)},

		{"03:04-05:00", time.Date(1, 1, 1, 3, 4, 0, 0, time.FixedZone("", -5*60*60))},
		{"03:04+05:00", time.Date(1, 1, 1, 3, 4, 0, 0, time.FixedZone("", 5*60*60))},
		{"03:04-0500", time.Date(1, 1, 1, 3, 4, 0, 0, time.FixedZone("", -5*60*60))},
		{"03:04+0500", time.Date(1, 1, 1, 3, 4, 0, 0, time.FixedZone("", 5*60*60))},
		{"03:04Z", time.Date(1, 1, 1, 3, 4, 0, 0, time.UTC)},
		{"03:04", time.Date(1, 1, 1, 3, 4, 0, 0, time.UTC)},

		{"03:04:05am", time.Date(1, 1, 1, 3, 4, 5, 0, time.UTC)},
		{"03:04:05pm", time.Date(1, 1, 1, 15, 4, 5, 0, time.UTC)},
		{"03:04AM", time.Date(1, 1, 1, 3, 4, 0, 0, time.UTC)},
		{"03:04PM", time.Date(1, 1, 1, 15, 4, 0, 0, time.UTC)},
		{"03a.m.", time.Date(1, 1, 1, 3, 0, 0, 0, time.UTC)},
		{"03p.m.", time.Date(1, 1, 1, 15, 0, 0, 0, time.UTC)},

		{"-05:00", time.Date(1, 1, 1, 0, 0, 0, 0, time.FixedZone("", -5*60*60))},
		{"+05:00", time.Date(1, 1, 1, 0, 0, 0, 0, time.FixedZone("", 5*60*60))},
		{"-0500", time.Date(1, 1, 1, 0, 0, 0, 0, time.FixedZone("", -5*60*60))},
		{"+0500", time.Date(1, 1, 1, 0, 0, 0, 0, time.FixedZone("", 5*60*60))},
		{"-05", time.Date(1, 1, 1, 0, 0, 0, 0, time.FixedZone("", -5*60*60))},
		{"+05", time.Date(1, 1, 1, 0, 0, 0, 0, time.FixedZone("", 5*60*60))},
		{"Z", time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC)},
	}

	for _, tt := range tests {
		var d datetime
		d.Parse(tt.input)
		if got, want := d.t, tt.want; !got.Equal(want) {
			t.Errorf("datetime.Parse(%v) returned %v, want %v", tt.input, got, want)
		}
	}
}

func Test_GetDateTimeValue(t *testing.T) {
	tests := []struct {
		html  string
		value *string
	}{
		{``, nil},
		{`<p><time class="value" datetime="2015-02-03T21:15:00-08:00"><time></p>`, ptr("2015-02-03 21:15:00-0800")},
		{`<p>
		    <time class="value" datetime="2015-02-03"></time>
		    <time class="value" datetime="21:15:00"></time>
		    <time class="value" datetime="-08:00"></time>
		  </p>`, ptr("2015-02-03 21:15:00-0800")},
	}

	for _, tt := range tests {
		n, err := parseNode(tt.html)
		if err != nil {
			t.Fatalf("Error parsing HTML: %v", err)
		}

		if got, want := getDateTimeValue(n), tt.value; !cmp.Equal(got, want) {
			t.Errorf("getDateTimeValue(%q) returned %v, want %v", tt.html, *got, *want)
		}
	}
}

func Test_ImplyEndDate(t *testing.T) {
	tests := []struct {
		description string
		start, end  []string
		want        []interface{}
	}{
		{
			"invalid dates",
			[]string{"foo"},
			[]string{"bar"},
			[]interface{}{"bar"},
		},
		{
			"single start and end date",
			[]string{"2006-01-02 03:04:05"},
			[]string{"01:02:03"},
			[]interface{}{"2006-01-02 01:02:03"},
		},
		{
			"single start and end date, end has date",
			[]string{"2006-01-02 03:04:05"},
			[]string{"2007-01-02 01:02:03"},
			[]interface{}{"2007-01-02 01:02:03"},
		},
		{
			"multiple start dates",
			[]string{"2006-01-02 03:04:05", "2007-01-02"},
			[]string{"01:02:03"},
			[]interface{}{"2006-01-02 01:02:03"},
		},
		{
			"multiple start dates, first with no date",
			[]string{"03:04:05", "2007-01-02"},
			[]string{"01:02:03"},
			[]interface{}{"2007-01-02 01:02:03"},
		},
		{
			"multiple start and end dates",
			[]string{"03:04:05", "2007-01-02"},
			[]string{"01:02:03", "2006-01-02 01:02:03"},
			[]interface{}{"2007-01-02 01:02:03", "2006-01-02 01:02:03"},
		},
	}

	for _, tt := range tests {
		item := &Microformat{Properties: map[string][]interface{}{}}
		for _, d := range tt.start {
			item.Properties["start"] = append(item.Properties["start"], d)
		}
		for _, d := range tt.end {
			item.Properties["end"] = append(item.Properties["end"], d)
		}

		implyEndDate(item)
		if got, want := item.Properties["end"], tt.want; !cmp.Equal(got, want) {
			t.Errorf("implyEndDate(%q, %q) returned %#v, want %#v", tt.start, tt.end, got, want)
		}
	}
}
