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

package model

// Bottle example
type Bottle struct {
	ID      int     `json:"id" example:"1"`
	Name    string  `json:"name" example:"bottle_name"`
	Account Account `json:"account"`
}

// BottlesAll example
func BottlesAll() ([]Bottle, error) {
	return bottles, nil
}

// BottleOne example
func BottleOne(id int) (*Bottle, error) {
	for _, v := range bottles {
		if id == v.ID {
			return &v, nil
		}
	}
	return nil, ErrNoRow
}

var bottles = []Bottle{
	{ID: 1, Name: "bottle_1", Account: Account{ID: 1, Name: "accout_1"}},
	{ID: 2, Name: "bottle_2", Account: Account{ID: 2, Name: "accout_2"}},
	{ID: 3, Name: "bottle_3", Account: Account{ID: 3, Name: "accout_3"}},
}
