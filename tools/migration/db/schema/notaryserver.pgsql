\c notaryserver;

CREATE TABLE "tuf_files" (
  "id" int PRIMARY KEY,
  "created_at" timestamp NULL DEFAULT NULL,
  "updated_at" timestamp NULL DEFAULT NULL,
  "deleted_at" timestamp NULL DEFAULT NULL,
  "gun" varchar(255) NOT NULL,
  "role" varchar(255) NOT NULL,
  "version" integer NOT NULL,
  "data" bytea NOT NULL,
  "sha256" char(64) DEFAULT NULL,
  UNIQUE ("gun","role","version")
);

CREATE INDEX tuf_files_sha256_idx ON tuf_files(sha256);

CREATE TABLE "change_category" (
    "category" VARCHAR(20) PRIMARY KEY
);

INSERT INTO "change_category" VALUES ('update'), ('deletion');

CREATE TABLE "changefeed" (
    "id" serial PRIMARY KEY,
    "created_at" timestamp DEFAULT CURRENT_TIMESTAMP,
    "gun" varchar(255) NOT NULL,
    "version" integer NOT NULL,
    "sha256" CHAR(64) DEFAULT NULL,
    "category" VARCHAR(20) NOT NULL DEFAULT 'update' REFERENCES "change_category"
);

CREATE INDEX "idx_changefeed_gun" ON "changefeed" ("gun");

INSERT INTO "changefeed" (
        "created_at",
        "gun",
        "version",
        "sha256" 
    ) (SELECT
        "created_at",
        "gun",
        "version",
        "sha256"
    FROM
        "tuf_files"
    WHERE
        "role" = 'timestamp'
    ORDER BY
        "created_at" ASC
);

CREATE TABLE "schema_migrations" (
  "version" int(11) NOT NULL,
  PRIMARY KEY ("version")
);

INSERT INTO "schema_migrations" VALUES (1),(2),(3),(4),(5);