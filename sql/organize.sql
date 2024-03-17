-- public.organize definition

-- Drop table

-- DROP TABLE public.organize;

CREATE TABLE public.organize (
	id SERIAL PRIMARY KEY,
	dep varchar(255) NOT NULL,
	org varchar(255) NULL
);
