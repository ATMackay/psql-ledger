CREATE TABLE "accounts" (
  "id" bigserial PRIMARY KEY,
  "username" varchar NOT NULL,
  "balance" bigint NOT NULL,
  "email" varchar,
  "created_at" timestamptz DEFAULT (now())
);

CREATE TABLE "transactions" (
  "id" bigserial PRIMARY KEY,
  "from_account" bigint,
  "to_account" bigint,
  "amount" bigint,
  "created_at" timestamptz DEFAULT (now())
);

CREATE INDEX ON "accounts" ("username");

CREATE INDEX ON "transactions" ("from");

CREATE INDEX ON "transactions" ("to");

ALTER TABLE "transactions" ADD FOREIGN KEY ("from") REFERENCES "accounts" ("id");

ALTER TABLE "transactions" ADD FOREIGN KEY ("to") REFERENCES "accounts" ("id");