// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Command signedurls creates a signed URL for a Cloud CDN endpoint with the
// given key.
package main

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"time"
)

// [START example]

// SignURL creates a signed URL for an endpoint on Cloud CDN.
//
// - url must start with "https://" and should not have the "Expires", "KeyName", or "Signature"
// query parameters.
// - key should be in raw form (not base64url-encoded) which is 16-bytes long.
// - keyName must match a key added to the backend service or bucket.
func SignURL(url, keyName string, key []byte, expiration time.Time) string {
	sep := "?"
	if strings.Contains(url, "?") {
		sep = "&"
	}
	url += sep
	url += fmt.Sprintf("Expires=%d", expiration.Unix())
	url += fmt.Sprintf("&KeyName=%s", keyName)

	mac := hmac.New(sha1.New, key)
	mac.Write([]byte(url))
	sig := base64.URLEncoding.EncodeToString(mac.Sum(nil))
	url += fmt.Sprintf("&Signature=%s", sig)
	return url
}

// SignURLWithPrefix creates a signed URL prefix for an endpoint on Cloud CDN.
// Prefixes allow access to any URL with the same prefix, and can be useful for
// granting access broader content without signing multiple URLs.
//
// - urlPrefix must start with "https://" and should not include query parameters.
// - key should be in raw form (not base64url-encoded) which is 16-bytes long.
// - keyName must match a key added to the backend service or bucket.
func SignURLWithPrefix(urlPrefix, keyName string, key []byte, expiration time.Time) (string, error) {
	if strings.Contains(urlPrefix, "?") {
		return "", fmt.Errorf("urlPrefix must not include query params: %s", urlPrefix)
	}

	encodedURLPrefix := base64.URLEncoding.EncodeToString([]byte(urlPrefix))
	input := fmt.Sprintf("URLPrefix=%s&Expires=%d&Keyname=%s",
		encodedURLPrefix,
		expiration.Unix(),
		keyName,
	)

	mac := hmac.New(sha1.New, key)
	mac.Write([]byte(input))
	sig := base64.URLEncoding.EncodeToString(mac.Sum(nil))

	signedValue := fmt.Sprintf("%s&Signature=%s",
		input,
		sig,
	)

	return signedValue, nil
}

// readKeyFile reads the base64url-encoded key file and decodes it.
func readKeyFile(path string) ([]byte, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read key file: %+v", err)
	}
	d := make([]byte, base64.URLEncoding.DecodedLen(len(b)))
	n, err := base64.URLEncoding.Decode(d, b)
	if err != nil {
		return nil, fmt.Errorf("failed to base64url decode: %+v", err)
	}
	return d[:n], nil
}

func main() {
	var keyPath string
	flag.StringVar(&keyPath, "key-file", "", "The path to a file containing the base64-encoded signing key")
	flag.Parse()

	key, err := readKeyFile(keyPath)
	if err != nil {
		log.Fatal(err)
	}

	url := SignURL("https://example.com", "my-key", key, time.Now().Add(time.Hour*24))
	fmt.Println(url)

	urlPrefix, err := SignURLWithPrefix("https://www.google.com/", "my-key", key, time.Unix(1549751401, 0))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(urlPrefix)
}

// [END example]
