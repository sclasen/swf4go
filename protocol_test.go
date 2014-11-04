package swf

import (
	"testing"
	"time"
)

func TestDateFormatting(t *testing.T) {
	swftime := SWFTime{time.Now()}

	b, err := swftime.MarshalJSON()

	if err != nil {
		t.Fatal(err)
	}

	testTime := new(SWFTime)
	err = testTime.UnmarshalJSON(b)

	if err != nil {
		t.Fatal(err)
	}

	if testTime.Time.Unix() != swftime.Time.Unix() {
		t.Fatal(swftime, testTime)
	}

}
