/*add robot account table*/
CREATE TABLE robot (
 id SERIAL PRIMARY KEY NOT NULL,
 name varchar(255),
 description varchar(1024),
 project_id int,
 expiresat bigint,
 disabled boolean DEFAULT false NOT NULL,
 creation_time timestamp default CURRENT_TIMESTAMP,
 update_time timestamp default CURRENT_TIMESTAMP,
 CONSTRAINT unique_robot UNIQUE (name, project_id)
);

CREATE TRIGGER robot_update_time_at_modtime BEFORE UPDATE ON robot FOR EACH ROW EXECUTE PROCEDURE update_update_time_at_column();

CREATE TABLE oidc_user_matadata (
 id SERIAL NOT NULL,
 user_id int NOT NULL,
 name varchar(255) NOT NULL,
 value varchar(255),
 creation_time timestamp default CURRENT_TIMESTAMP,
 update_time timestamp default CURRENT_TIMESTAMP,
 PRIMARY KEY (id),
 CONSTRAINT unique_user_id_and_name UNIQUE (user_id,name),
 FOREIGN KEY (user_id) REFERENCES harbor_user(user_id)
);

CREATE TRIGGER odic_user_metadata_update_time_at_modtime BEFORE UPDATE ON oidc_user_matadata FOR EACH ROW EXECUTE PROCEDURE update_update_time_at_column();

/*add master role*/
INSERT INTO role (role_code, name) VALUES ('DRWS', 'master');

/*delete replication jobs whose policy has been marked as "deleted"*/
DELETE FROM replication_job AS j
USING replication_policy AS p
WHERE j.policy_id = p.id AND p.deleted = TRUE;

/*delete replication policy which has been marked as "deleted"*/
DELETE FROM replication_policy AS p
WHERE p.deleted = TRUE;