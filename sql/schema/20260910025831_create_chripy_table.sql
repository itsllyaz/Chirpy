-- +goose Up
-- +goose StatementBegin
CREATE TABLE "chirpy" (
  "id" uuid PRIMARY KEY,
  "body" text,
	"created_at" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	"updated_at" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "user_id" uuid
);

ALTER TABLE "chirpy" 
ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id")
ON DELETE CASCADE
DEFERRABLE INITIALLY IMMEDIATE; 
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS chirpy; 

-- +goose StatementEnd


