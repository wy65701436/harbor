ALTER TABLE schedule ADD COLUMN IF NOT EXISTS cron_type varchar(64);
ALTER TABLE robot ADD COLUMN IF NOT EXISTS secret varchar(2048);

CREATE TABLE  IF NOT EXISTS role_permission (
 id SERIAL PRIMARY KEY NOT NULL,
 role_type varchar(255) NOT NULL,
 role_id int NOT NULL,
 rbac_policy_id int NOT NULL,
 creation_time timestamp default CURRENT_TIMESTAMP,
 update_time timestamp default CURRENT_TIMESTAMP,
 CONSTRAINT unique_role_permission UNIQUE (role_type, role_id, rbac_policy_id)
);

CREATE TABLE  IF NOT EXISTS rbac_policy (
 id SERIAL PRIMARY KEY NOT NULL,
 /*
  scope:
   system level: /system
   project level: /project/{id}
   all project: /project/ *
  */
 scope varchar(255),
 resource varchar(255),
 action varchar(255),
 effect varchar(255),
 creation_time timestamp default CURRENT_TIMESTAMP,
 update_time timestamp default CURRENT_TIMESTAMP,
 CONSTRAINT unique_rbac_policy UNIQUE (scope, resource, action, effect)
);

