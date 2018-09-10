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

*** Settings ***
Documentation  Harbor BATs
Resource  ../../resources/APITest-Util.robot
Library  OperatingSystem
Library  String
Library  Collections
Library  requests
Library  Process
Suite Setup  Setup API Test
Default Tags  E2E

*** Test Cases ***
Test Case - Run LDAP Group Related API Test
    Harbor API Test  ./tests/apitests/python/test_ldap_admin_role.py
    Harbor API Test  ./tests/apitests/python/test_user_group.py
    Harbor API Test  ./tests/apitests/python/test_assign_role_to_ldap_group.py