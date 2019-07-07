// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package parser

const mysqlSQL = `CREATE TABLE gp_options (
  key varchar(50) NOT NULL,
  value longtext NOT NULL,
  group varchar(50) NOT NULL,
  type varchar(20) NOT NULL,
  PRIMARY KEY (key),
  UNIQUE KEY u_group_key (key,group)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci`
