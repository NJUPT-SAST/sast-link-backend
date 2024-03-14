-- public."admin" definition

-- Drop table

-- DROP TABLE public."admin";

CREATE TABLE public."admin" (
	id serial4 NOT NULL,
	created_at timestamp NOT NULL DEFAULT now(),
	user_id varchar(255) NOT NULL,
	CONSTRAINT admin_pkey PRIMARY KEY (id)
);
