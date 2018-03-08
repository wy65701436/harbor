// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package test

import (
	"os"
	"strconv"

	"github.com/vmware/harbor/src/common/dao"
	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/utils/log"
)

// InitDatabaseFromEnv is used to initialize database for testing
func InitDatabaseFromEnv() {
	dbHost := os.Getenv("POSTGRES_HOST")
	if len(dbHost) == 0 {
		log.Fatalf("environment variable POSTGRES_HOST is not set")
	}
	dbUser := os.Getenv("POSTGRES_USR")
	if len(dbUser) == 0 {
		log.Fatalf("environment variable POSTGRES_USR is not set")
	}
	dbPortStr := os.Getenv("POSTGRES_PORT")
	if len(dbPortStr) == 0 {
		log.Fatalf("environment variable POSTGRES_PORT is not set")
	}
	dbPort, err := strconv.Atoi(dbPortStr)
	if err != nil {
		log.Fatalf("invalid POSTGRES_PORT: %v", err)
	}

	dbPassword := os.Getenv("POSTGRES_PWD")
	dbDatabase := os.Getenv("POSTGRES_DATABASE")
	if len(dbDatabase) == 0 {
		log.Fatalf("environment variable POSTGRES_DATABASE is not set")
	}

	database := &models.Database{
		Type: "postgres",
		PostGreSQL: &models.PostGreSQL{
			Host:     dbHost,
			Port:     dbPort,
			Username: dbUser,
			Password: dbPassword,
			Database: dbDatabase,
		},
	}

	log.Infof("POSTGRES_HOST: %s, POSTGRES_USR: %s, POSTGRES_PORT: %d, POSTGRES_PWD: %s\n", dbHost, dbUser, dbPort, dbPassword)

	if err := dao.InitDatabase(database); err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}
}
