-- public.carrer_records definition

-- Drop table

-- DROP TABLE public.carrer_records;

CREATE TABLE public.carrer_records (
	id serial4 NOT NULL,
	user_id int4 NOT NULL, -- 与user表映射，表示某个用户的生涯记录
	org_id int2 NOT NULL, -- 与orgnize表映射，表示用户该届所在的组织
	grade int2 NOT NULL, -- 表示某一届（如：2023届）
	is_delete bool NOT NULL, -- 假删
	"position" varchar(2) NULL, -- 包括:部员、讲师、组长、部长、主席
	CONSTRAINT carrer_records_pkey PRIMARY KEY (id)
);

-- Column comments

COMMENT ON COLUMN public.carrer_records.user_id IS '与user表映射，表示某个用户的生涯记录';
COMMENT ON COLUMN public.carrer_records.org_id IS '与orgnize表映射，表示用户该届所在的组织';
COMMENT ON COLUMN public.carrer_records.grade IS '表示某一届（如：2023届）';
COMMENT ON COLUMN public.carrer_records.is_delete IS '假删';
COMMENT ON COLUMN public.carrer_records."position" IS '包括:部员、讲师、组长、部长、主席';
