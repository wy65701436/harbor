ALTER TABLE blob ADD COLUMN update_time timestamp default CURRENT_TIMESTAMP;
ALTER TABLE blob ADD COLUMN status varchar(255);