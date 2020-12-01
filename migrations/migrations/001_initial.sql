-- Write your migrate up statements here

{{ template "migrations/shared/trigger_set_timestamp.sql" . }}


create table item_state (
  id int not null UNIQUE,
  type text not null UNIQUE
  );

insert into item_state (id, type) values(0, 'disabled');
insert into item_state (id, type) values(1, 'valid');

create table items (
  uuid uuid PRIMARY KEY,
  title TEXT NOT NULL,
  description TEXT,
  language_code varchar(2) NOT NULL,
  publication_uuid uuid NOT NULL,
  published_date timestamptz NOT NULL,
  content TEXT,
  url TEXT,
  state_id int NOT NULL REFERENCES item_state(id),
  created_at timestamptz NOT NULL DEFAULT NOW(),
  modified_at timestamptz NOT NULL DEFAULT NOW()
);

CREATE TRIGGER set_timestamp BEFORE UPDATE ON "items" FOR EACH ROW EXECUTE PROCEDURE trigger_set_timestamp();

---- create above / drop below ----

DROP trigger set_timestamp ON "items";

DROP FUNCTION trigger_set_timestamp;

DROP TABLE "items";
DROP TABLE "world_languages"
DROP TABLE "item_state"

-- Write your migrate down statements here. If this migration is irreversible
-- Then delete the separator line above.
