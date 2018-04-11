#ifndef _13HOSTS_PROTON_COMMON_H
#define _13HOSTS_PROTON_COMMON_H 1
/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 *
 */

// An ordered list of server addresses for reconnect
#define MAX_HOSTS 10
typedef struct
{
    int current;
    int count;  // # of hosts
    char *hosts[MAX_HOSTS];
} hosts_t;

int hosts_init(hosts_t *hosts, char *list);
const char *hosts_get(hosts_t *hosts);
void fatal(const char *str);

#endif


