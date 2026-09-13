-- +goose Up
ALTER TABLE "chirpy"
ALTER COLUMN "body" SET NOT NULL, 
ALTER COLUMN "user_id" SET NOT NULL;

-- +goose Down
ALTER TABLE "chirpy" 
ALTER COLUMN "body" DROP NOT NULL, 
ALTER COLUMN "user_id" DROP NOT NULL; 

