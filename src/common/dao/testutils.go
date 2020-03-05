// Copyright Project Harbor Authors
//
// licensed under the apache license, version 2.0 (the "license");
// you may not use this file except in compliance with the license.
// you may obtain a copy of the license at
//
//    http://www.apache.org/licenses/license-2.0
//
// unless required by applicable law or agreed to in writing, software
// distributed under the license is distributed on an "as is" basis,
// without warranties or conditions of any kind, either express or implied.
// see the license for the specific language governing permissions and
// limitations under the license.

package dao

import (
	"fmt"
	"os"
	"strconv"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/log"
)

var defaultRegistered = false

// PrepareTestForMySQL is for test only.
func PrepareTestForMySQL() {
}

// PrepareTestForSQLite is for test only.
func PrepareTestForSQLite() {
}

// PrepareTestForPostgresSQL is for test only.
func PrepareTestForPostgresSQL() {
	z

	database := &models.Database{
		Type: "postgresql",
		PostGreSQL: &models.PostGreSQL{
			Host:     dbHost,
			Port:     dbPort,
			Username: dbUser,
			Password: dbPassword,
			Database: dbDatabase,
		},
	}

	log.Infof("POSTGRES_HOST: %s, POSTGRES_USR: %s, POSTGRES_PORT: %d, POSTGRES_PWD: %s\n", dbHost, dbUser, dbPort, dbPassword)
	initDatabaseForTest(database)
}

func initDatabaseForTest(db *models.Database) {
	database, err := getDatabase(db)
	if err != nil {
		panic(err)
	}

	log.Infof("initializing database: %s", database.String())

	alias := database.Name()
	if !defaultRegistered {
		defaultRegistered = true
		alias = "default"
	}
	if err := database.Register(alias); err != nil {
		panic(err)
	}
	if err := database.UpgradeSchema(); err != nil {
		panic(err)
	}

	if alias != "default" {
		if err = globalOrm.Using(alias); err != nil {
			log.Fatalf("failed to create new orm: %v", err)
		}
	}
}

// PrepareTestData -- Clean and Create data
func PrepareTestData(clearSqls []string, initSqls []string) {
	o := GetOrmer()

	for _, sql := range clearSqls {
		fmt.Printf("Exec sql:%v\n", sql)
		_, err := o.Raw(sql).Exec()
		if err != nil {
			fmt.Printf("failed to clear database, sql:%v, error: %v", sql, err)
		}
	}

	for _, sql := range initSqls {
		_, err := o.Raw(sql).Exec()
		if err != nil {
			fmt.Printf("failed to init database, sql:%v, error: %v", sql, err)
		}
	}
}

// ExecuteBatchSQL ...
func ExecuteBatchSQL(sqls []string) {
	o := GetOrmer()

	for _, sql := range sqls {
		fmt.Printf("Exec sql:%v\n", sql)
		_, err := o.Raw(sql).Exec()
		if err != nil {
			fmt.Printf("failed to execute batch sql, sql:%v, error: %v", sql, err)
		}
	}
}

// ArrayEqual ...
func ArrayEqual(arrayA, arrayB []int) bool {
	if len(arrayA) != len(arrayB) {
		return false
	}
	size := len(arrayA)
	for i := 0; i < size; i++ {
		if arrayA[i] != arrayB[i] {
			return false
		}
	}
	return true
}

// ClearHTTPAuthProxyUsers remove the records from harbor_users to delete all user imported via
// HTTP Auth Proxy
func ClearHTTPAuthProxyUsers() error {
	o := GetOrmer()
	sql := "DELETE FROM harbor_user WHERE comment='By Authproxy'"
	_, err := o.Raw(sql).Exec()
	return err
}
