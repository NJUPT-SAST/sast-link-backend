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
-- Name: oauth2_info id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.oauth2_info ALTER COLUMN id SET DEFAULT nextval('public.oauth2_info_id_seq'::regclass);


--
-- Name: oauth2_info oauth2_info_unique; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.oauth2_info
    ADD CONSTRAINT oauth2_info_unique UNIQUE (client, user_id);


--
-- PostgreSQL database dump complete
--

