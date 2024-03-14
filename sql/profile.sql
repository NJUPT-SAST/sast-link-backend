-- public.profile definition

-- Drop table

-- DROP TABLE public.profile;

CREATE TABLE public.profile (
	id serial4 NOT NULL,
	user_id int4 NOT NULL, -- 与user表映射
	nickname varchar(255) NOT NULL, -- 昵称
	org_id int2 NOT NULL, -- 对应部门和组的信息（现在的职位，历史职位的信息在carrer_records中）
	bio varchar(255) NULL, -- 自我介绍
	email varchar(255) NOT NULL, -- 邮箱(默认展示)
	badge json NULL, -- 纪念卡
	link _varchar NULL, -- 个人链接（包括自己b站、博客、GitHub等账号链接）
	avatar varchar(255) NULL, -- 头像（存储oss链接）
	is_deleted bool NOT NULL, -- 假删
	hide _varchar NULL, -- 选择隐藏的信息
	CONSTRAINT profile_pkey PRIMARY KEY (id)
);

-- Column comments

COMMENT ON COLUMN public.profile.user_id IS '与user表映射';
COMMENT ON COLUMN public.profile.nickname IS '昵称';
COMMENT ON COLUMN public.profile.org_id IS '对应部门和组的信息（现在的职位，历史职位的信息在carrer_records中）';
COMMENT ON COLUMN public.profile.bio IS '自我介绍';
COMMENT ON COLUMN public.profile.email IS '邮箱(默认展示)';
COMMENT ON COLUMN public.profile.badge IS '纪念卡';
COMMENT ON COLUMN public.profile.link IS '个人链接（包括自己b站、博客、GitHub等账号链接）';
COMMENT ON COLUMN public.profile.avatar IS '头像（存储oss链接）';
COMMENT ON COLUMN public.profile.is_deleted IS '假删';
COMMENT ON COLUMN public.profile.hide IS '选择隐藏的信息';
