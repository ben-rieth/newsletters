\restrict gUNWdhsMD1pkhvrQB5YWBgrjwPP8jGMwj84HYGo07jgAgfNOQzT0QxNmDtwogyi

-- Dumped from database version 16.13 (Ubuntu 16.13-0ubuntu0.24.04.1)
-- Dumped by pg_dump version 16.13 (Ubuntu 16.13-0ubuntu0.24.04.1)

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

--
-- Name: frequency; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.frequency AS ENUM (
    'monthly',
    'weekly',
    'daily'
);


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: newsletters; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.newsletters (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name text NOT NULL,
    frequency public.frequency NOT NULL,
    send_day integer DEFAULT 0 NOT NULL,
    send_hour integer NOT NULL,
    send_minute integer DEFAULT 0 NOT NULL,
    last_sent_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: schema_migrations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.schema_migrations (
    version character varying NOT NULL
);


--
-- Name: newsletters newsletters_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.newsletters
    ADD CONSTRAINT newsletters_pkey PRIMARY KEY (id);


--
-- Name: schema_migrations schema_migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.schema_migrations
    ADD CONSTRAINT schema_migrations_pkey PRIMARY KEY (version);


--
-- PostgreSQL database dump complete
--

\unrestrict gUNWdhsMD1pkhvrQB5YWBgrjwPP8jGMwj84HYGo07jgAgfNOQzT0QxNmDtwogyi


--
-- Dbmate schema migrations
--

INSERT INTO public.schema_migrations (version) VALUES
    ('20260329010424');
