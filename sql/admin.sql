-- public."admin" definition

-- Drop table

-- DROP TABLE public."admin";

CREATE TABLE public."admin" (
	id SERIAL PRIMARY KEY,
	created_at timestamp NOT NULL DEFAULT now(),
	user_id varchar(255) NOT NULL
);
