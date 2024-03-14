-- public."user" definition

-- Drop table

-- DROP TABLE public."user";

CREATE TABLE public."user" (
	id serial4 NOT NULL,
	created_at timestamp NOT NULL DEFAULT now(),
	email varchar(255) NOT NULL,
	uid varchar(255) NOT NULL,
	qq_id varchar(255) NULL,
	lark_id varchar(255) NULL,
	github_id varchar(255) NULL,
	wechat_id varchar(255) NULL,
	is_deleted bool NOT NULL,
	"password" varchar(255) NOT NULL,
	CONSTRAINT user_pkey PRIMARY KEY (id)
);
