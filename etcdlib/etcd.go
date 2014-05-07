// Copyright (c) 2014 The SkyDNS Authors. All rights reserved.
// Use of this source code is governed by The MIT License (MIT) that can be
// found in the LICENSE file.

package etcdlib

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/DimShadoWWW/integrator/dockerlib"
	"github.com/coreos/go-etcd/etcd"
)

func RegisterDNS(client *etcd.Client, container dockerlib.APIContainers, region string) error {
	value, err := json.Marshal("{}")
	if err != nil {
		return err
	}

	key := []byte(value)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(container.ID))
	t := base64.StdEncoding.EncodeToString(h.Sum(nil))

	hash := sha256.New()
	hash.Write([]byte(container.ID))
	r, err := client.Set("/skydns/local/"+container.Hostname+"/"+t,
		string(value),
		uint64(0))
	if err != nil {
		return err
	}
	fmt.Println(r)
	return nil
}
