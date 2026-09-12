-- name: CreateChirpy :one
INSERT INTO chirpy (id, body,  created_at, updated_at, user_id)
VALUES (
  gen_random_uuid(), 
	$1, 
	NOW(), 
	NOW(), 
	$2
	
)
RETURNING *;

-- name: GetChirpys :many 
SELECT * FROM chirpy; 

-- name: GetChirpyByID :one 
SELECT * FROM chirpy WHERE id = $1;
