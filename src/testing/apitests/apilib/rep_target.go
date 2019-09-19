/*
 * Harbor API
 *
 * These APIs provide services for manipulating Harbor project.
 *
 * OpenAPI spec version: 0.3.0
 *
 * Generated by: https://github.com/swagger-api/swagger-codegen.git
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package apilib

type RepTarget struct {

	// The target ID.
	Id int64 `json:"id,omitempty"`

	// The target address URL string.
	Endpoint string `json:"endpoint,omitempty"`

	// The target name.
	Name string `json:"name,omitempty"`

	// The target server username.
	Username string `json:"username,omitempty"`

	// The target server password.
	Password string `json:"password,omitempty"`

	// Reserved field.
	Type_ int32 `json:"type,omitempty"`

	// The create time of the rule.
	CreationTime string `json:"creation_time,omitempty"`

	// The update time of the rule.
	UpdateTime string `json:"update_time,omitempty"`
}
