--
-- PostgreSQL database dump
--

-- Dumped from database version 15.4 (Debian 15.4-1.pgdg120+1)
-- Dumped by pg_dump version 15.4 (Debian 15.4-1.pgdg120+1)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: admin; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.admin (
    id integer NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    user_id character varying(255) NOT NULL
);


ALTER TABLE public.admin OWNER TO postgres;

--
-- Name: admin_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.admin_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.admin_id_seq OWNER TO postgres;

--
-- Name: admin_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.admin_id_seq OWNED BY public.admin.id;


--
-- Name: carrer_records; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.carrer_records (
    id integer NOT NULL,
    user_id integer NOT NULL,
    org_id smallint NOT NULL,
    grade smallint NOT NULL,
    is_delete boolean NOT NULL,
    "position" character varying(2)
);


ALTER TABLE public.carrer_records OWNER TO postgres;

--
-- Name: COLUMN carrer_records.user_id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.carrer_records.user_id IS '与user表映射，表示某个用户的生涯记录';


--
-- Name: COLUMN carrer_records.org_id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.carrer_records.org_id IS '与orgnize表映射，表示用户该届所在的组织';


--
-- Name: COLUMN carrer_records.grade; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.carrer_records.grade IS '表示某一届（如：2023届）';


--
-- Name: COLUMN carrer_records.is_delete; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.carrer_records.is_delete IS '假删';


--
-- Name: COLUMN carrer_records."position"; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.carrer_records."position" IS '包括:部员、讲师、组长、部长、主席';


--
-- Name: carrer_records_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.carrer_records_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.carrer_records_id_seq OWNER TO postgres;

--
-- Name: carrer_records_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.carrer_records_id_seq OWNED BY public.carrer_records.id;


--
-- Name: organize; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.organize (
    id integer NOT NULL,
    dep character varying(255) NOT NULL,
    org character varying(255)
);


ALTER TABLE public.organize OWNER TO postgres;

--
-- Name: department_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.department_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.department_id_seq OWNER TO postgres;

--
-- Name: department_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.department_id_seq OWNED BY public.organize.id;


--
-- Name: oauth2_clients; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.oauth2_clients (
    id text NOT NULL,
    secret text NOT NULL,
    domain text NOT NULL,
    data jsonb NOT NULL
);


ALTER TABLE public.oauth2_clients OWNER TO postgres;

--
-- Name: oauth2_info; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.oauth2_info (
    id integer NOT NULL,
    client character varying NOT NULL,
    info jsonb[],
    oauth_user_id character varying NOT NULL,
    user_id character varying NOT NULL
);


ALTER TABLE public.oauth2_info OWNER TO postgres;

--
-- Name: COLUMN oauth2_info.client; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.oauth2_info.client IS 'oauth client, eg. Feishu';


--
-- Name: COLUMN oauth2_info.oauth_user_id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.oauth2_info.oauth_user_id IS 'user id, eg. github user id.';


--
-- Name: COLUMN oauth2_info.user_id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.oauth2_info.user_id IS 'user id, eg. B21010101';


--
-- Name: oauth2_info_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.oauth2_info_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.oauth2_info_id_seq OWNER TO postgres;

--
-- Name: oauth2_info_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.oauth2_info_id_seq OWNED BY public.oauth2_info.id;


--
-- Name: oauth2_tokens; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.oauth2_tokens (
    id bigint NOT NULL,
    created_at timestamp with time zone NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    code text NOT NULL,
    access text NOT NULL,
    refresh text NOT NULL,
    data jsonb NOT NULL
);


ALTER TABLE public.oauth2_tokens OWNER TO postgres;

--
-- Name: oauth2_tokens_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.oauth2_tokens_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.oauth2_tokens_id_seq OWNER TO postgres;

--
-- Name: oauth2_tokens_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.oauth2_tokens_id_seq OWNED BY public.oauth2_tokens.id;


--
-- Name: profile; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.profile (
    id integer NOT NULL,
    user_id integer NOT NULL,
    nickname character varying(255) NOT NULL,
    org_id smallint NOT NULL,
    bio character varying(255),
    email character varying(255) NOT NULL,
    badge json,
    link character varying[],
    avatar character varying(255),
    is_deleted boolean NOT NULL,
    hide character varying[]
);


ALTER TABLE public.profile OWNER TO postgres;

--
-- Name: COLUMN profile.user_id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.profile.user_id IS '与user表映射';


--
-- Name: COLUMN profile.nickname; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.profile.nickname IS '昵称';


--
-- Name: COLUMN profile.org_id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.profile.org_id IS '对应部门和组的信息（现在的职位，历史职位的信息在carrer_records中）';


--
-- Name: COLUMN profile.bio; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.profile.bio IS '自我介绍';


--
-- Name: COLUMN profile.email; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.profile.email IS '邮箱(默认展示)';


--
-- Name: COLUMN profile.badge; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.profile.badge IS '纪念卡';


--
-- Name: COLUMN profile.link; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.profile.link IS '个人链接（包括自己b站、博客、GitHub等账号链接）';


--
-- Name: COLUMN profile.avatar; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.profile.avatar IS '头像（存储oss链接）';


--
-- Name: COLUMN profile.is_deleted; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.profile.is_deleted IS '假删';


--
-- Name: COLUMN profile.hide; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.profile.hide IS '选择隐藏的信息';


--
-- Name: profile_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.profile_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.profile_id_seq OWNER TO postgres;

--
-- Name: profile_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.profile_id_seq OWNED BY public.profile.id;


--
-- Name: user; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public."user" (
    id integer NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    email character varying(255) NOT NULL,
    uid character varying(255) NOT NULL,
    qq_id character varying(255),
    lark_id character varying(255),
    github_id character varying(255),
    wechat_id character varying(255),
    is_deleted boolean NOT NULL,
    password character varying(255) NOT NULL
);


ALTER TABLE public."user" OWNER TO postgres;

--
-- Name: user_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.user_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.user_id_seq OWNER TO postgres;

--
-- Name: user_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.user_id_seq OWNED BY public."user".id;


--
-- Name: admin id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.admin ALTER COLUMN id SET DEFAULT nextval('public.admin_id_seq'::regclass);


--
-- Name: carrer_records id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.carrer_records ALTER COLUMN id SET DEFAULT nextval('public.carrer_records_id_seq'::regclass);


--
-- Name: oauth2_info id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.oauth2_info ALTER COLUMN id SET DEFAULT nextval('public.oauth2_info_id_seq'::regclass);


--
-- Name: oauth2_tokens id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.oauth2_tokens ALTER COLUMN id SET DEFAULT nextval('public.oauth2_tokens_id_seq'::regclass);


--
-- Name: organize id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.organize ALTER COLUMN id SET DEFAULT nextval('public.department_id_seq'::regclass);


--
-- Name: profile id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.profile ALTER COLUMN id SET DEFAULT nextval('public.profile_id_seq'::regclass);


--
-- Name: user id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public."user" ALTER COLUMN id SET DEFAULT nextval('public.user_id_seq'::regclass);


--
-- Name: admin admin_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.admin
    ADD CONSTRAINT admin_pkey PRIMARY KEY (id);


--
-- Name: carrer_records carrer_records_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.carrer_records
    ADD CONSTRAINT carrer_records_pkey PRIMARY KEY (id);


--
-- Name: organize department_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.organize
    ADD CONSTRAINT department_pkey PRIMARY KEY (id);


--
-- Name: oauth2_clients oauth2_clients_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.oauth2_clients
    ADD CONSTRAINT oauth2_clients_pkey PRIMARY KEY (id);


--
-- Name: oauth2_info oauth2_info_unique; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.oauth2_info
    ADD CONSTRAINT oauth2_info_unique UNIQUE (client, user_id);


--
-- Name: oauth2_tokens oauth2_tokens_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.oauth2_tokens
    ADD CONSTRAINT oauth2_tokens_pkey PRIMARY KEY (id);


--
-- Name: profile profile_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.profile
    ADD CONSTRAINT profile_pkey PRIMARY KEY (id);


--
-- Name: user user_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public."user"
    ADD CONSTRAINT user_pkey PRIMARY KEY (id);


--
-- Name: user user_un; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public."user"
    ADD CONSTRAINT user_un UNIQUE (uid, id);


--
-- Name: idx_oauth2_tokens_access; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_oauth2_tokens_access ON public.oauth2_tokens USING btree (access);


--
-- Name: idx_oauth2_tokens_code; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_oauth2_tokens_code ON public.oauth2_tokens USING btree (code);


--
-- Name: idx_oauth2_tokens_expires_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_oauth2_tokens_expires_at ON public.oauth2_tokens USING btree (expires_at);


--
-- Name: idx_oauth2_tokens_refresh; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_oauth2_tokens_refresh ON public.oauth2_tokens USING btree (refresh);


--
-- PostgreSQL database dump complete
--

