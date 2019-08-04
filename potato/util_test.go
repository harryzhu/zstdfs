package potato

import (
	"testing"
)

func TestZip(t *testing.T) {

}

func TestUnzip(t *testing.T) {
	str := "OK"
	str_zipped := Zip([]byte(str))
	str_zipped_unzipped := Unzip(str_zipped)
	if string(str_zipped_unzipped) != str {
		t.Log("Func Zip/Unzip failed the test.")
		t.FailNow()
	}
}

func TestByteMD5(t *testing.T) {
	str := "OK"
	str_md5 := "e0aa021e21dddbd6d8cecec71e9cf564"
	if ByteMD5([]byte(str)) != str_md5 {
		t.Log("Func ByteMD5 failed the test.")
		t.FailNow()
	}
}

func TestByteSHA256(t *testing.T) {
	str := "OK"
	str_sha256 := "565339bc4d33d72817b583024112eb7f5cdf3e5eef0252d6ec1b9c9a94e12bb3"
	if ByteSHA256([]byte(str)) != str_sha256 {
		t.Log("Func ByteSHA256 failed the test.")
		t.FailNow()
	}
}
