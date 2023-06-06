/*
Copyright 2023-2023 VMware Inc.
SPDX-License-Identifier: Apache-2.0

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package persist

import (
	"errors"

	_ "k8s.io/utils/lru"
)

var (
	// TODO: change to persist store
	collections map[string]Store
)

func init() {
	collections = map[string]Store{}
}

type Store interface {
	Put(id string, data interface{})
	Get(id string) (interface{}, error)
}

type storeImpl struct {
	data map[string]interface{}
}

func Collection(name string) Store {
	s, ok := collections[name]
	if !ok {
		s = &storeImpl{data: map[string]interface{}{}}
		collections[name] = s
	}
	return s
}

func (s *storeImpl) Put(id string, data interface{}) {
	s.data[id] = data
}

func (s *storeImpl) Get(id string) (interface{}, error) {
	v, exist := s.data[id]
	if exist {
		return v, nil
	}
	return nil, errors.New("Item not found: " + id)
}
