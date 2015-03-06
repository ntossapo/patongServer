package server_test

import(
	"testing"
	"net/http"
	"net/url"
	"strconv"
	"io/ioutil"
)

func TestGetBestpath(t *testing.T){

	res,_ := http.PostForm("http://127.0.0.1:8080/route",
	url.Values{"oriLat": {"7.8923159"}, "oriLong": {"98.3691285"}, "destLat":{"7.9074752"}, "destLong":{"98.3506281"}})
	content, _ := ioutil.ReadAll(res.Body)
	t.Log(content)
	if res.StatusCode  != 200 {
		t.Errorf("expect 200 but was %d", res.StatusCode)
	}
}

func FloatToString(inputFloat float64) string{
	return strconv.FormatFloat(inputFloat, 'f', 6, 64)
}

func TestFloatToString(t *testing.T){
	if FloatToString(9.5) != "9.500000" {
		t.Errorf("expected 9.5 but was %s", FloatToString(9.5))
	}
}
