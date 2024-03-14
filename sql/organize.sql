-- public.organize definition

-- Drop table

-- DROP TABLE public.organize;

CREATE TABLE public.organize (
	id int4 NOT NULL DEFAULT nextval('department_id_seq'::regclass),
	dep varchar(255) NOT NULL,
	org varchar(255) NULL,
	CONSTRAINT department_pkey PRIMARY KEY (id)
);
