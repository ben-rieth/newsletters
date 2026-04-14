\restrict 4b4LclgcPQg6rDm8eT15swLJz5mQFDmUqSRzeeOjpgBNyKAnXY52ShEyMdlN5EG

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
-- Name: feedurlsource; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.feedurlsource AS ENUM (
    'canonical',
    'user_submitted',
    'in_feed_response',
    'unknown'
);


--
-- Name: frequency; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.frequency AS ENUM (
    'monthly',
    'weekly',
    'daily'
);


--
-- Name: itemstate; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.itemstate AS ENUM (
    'read',
    'unread'
);


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: app_user; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.app_user (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    email text NOT NULL,
    password text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: feed; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.feed (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    title text NOT NULL,
    description text NOT NULL,
    url text NOT NULL,
    last_retrieved_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: feed_item; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.feed_item (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    feed_id uuid NOT NULL,
    title text NOT NULL,
    url text NOT NULL,
    publish_date timestamp with time zone NOT NULL,
    retrieved_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: feed_url; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.feed_url (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    feed_id uuid NOT NULL,
    url text NOT NULL,
    source public.feedurlsource NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: newsletter; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.newsletter (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name text NOT NULL,
    frequency public.frequency NOT NULL,
    send_day integer DEFAULT 0 NOT NULL,
    send_hour integer NOT NULL,
    send_minute integer DEFAULT 0 NOT NULL,
    send_timezone text DEFAULT 'UTC'::text NOT NULL,
    last_sent_at timestamp with time zone,
    next_send_time timestamp with time zone NOT NULL,
    user_id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: newsletter_feed; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.newsletter_feed (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    newsletter_id uuid NOT NULL,
    feed_id uuid NOT NULL,
    user_id uuid NOT NULL,
    alias text DEFAULT ''::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: newsletter_feed_item_status; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.newsletter_feed_item_status (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    newsletter_feed_id uuid NOT NULL,
    item_id uuid NOT NULL,
    user_id uuid NOT NULL,
    state public.itemstate DEFAULT 'unread'::public.itemstate NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: refresh_token; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.refresh_token (
    id bigint NOT NULL,
    token text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    user_id uuid NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    revoked_at timestamp with time zone
);


--
-- Name: refresh_token_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.refresh_token_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: refresh_token_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.refresh_token_id_seq OWNED BY public.refresh_token.id;


--
-- Name: schema_migrations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.schema_migrations (
    version character varying NOT NULL
);


--
-- Name: refresh_token id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.refresh_token ALTER COLUMN id SET DEFAULT nextval('public.refresh_token_id_seq'::regclass);


--
-- Name: app_user app_user_email_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.app_user
    ADD CONSTRAINT app_user_email_key UNIQUE (email);


--
-- Name: app_user app_user_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.app_user
    ADD CONSTRAINT app_user_pkey PRIMARY KEY (id);


--
-- Name: feed_item feed_item_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.feed_item
    ADD CONSTRAINT feed_item_pkey PRIMARY KEY (id);


--
-- Name: feed feed_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.feed
    ADD CONSTRAINT feed_pkey PRIMARY KEY (id);


--
-- Name: feed feed_url_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.feed
    ADD CONSTRAINT feed_url_key UNIQUE (url);


--
-- Name: feed_url feed_url_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.feed_url
    ADD CONSTRAINT feed_url_pkey PRIMARY KEY (id);


--
-- Name: feed_url feed_url_url_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.feed_url
    ADD CONSTRAINT feed_url_url_key UNIQUE (url);


--
-- Name: newsletter_feed_item_status newsletter_feed_item_status_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.newsletter_feed_item_status
    ADD CONSTRAINT newsletter_feed_item_status_pkey PRIMARY KEY (id);


--
-- Name: newsletter_feed newsletter_feed_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.newsletter_feed
    ADD CONSTRAINT newsletter_feed_pkey PRIMARY KEY (id);


--
-- Name: newsletter newsletter_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.newsletter
    ADD CONSTRAINT newsletter_pkey PRIMARY KEY (id);


--
-- Name: refresh_token refresh_token_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.refresh_token
    ADD CONSTRAINT refresh_token_pkey PRIMARY KEY (id);


--
-- Name: refresh_token refresh_token_token_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.refresh_token
    ADD CONSTRAINT refresh_token_token_key UNIQUE (token);


--
-- Name: schema_migrations schema_migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.schema_migrations
    ADD CONSTRAINT schema_migrations_pkey PRIMARY KEY (version);


--
-- Name: feed_item feed_item_feed_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.feed_item
    ADD CONSTRAINT feed_item_feed_id_fkey FOREIGN KEY (feed_id) REFERENCES public.feed(id) ON UPDATE RESTRICT ON DELETE RESTRICT;


--
-- Name: feed_url feed_url_feed_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.feed_url
    ADD CONSTRAINT feed_url_feed_id_fkey FOREIGN KEY (feed_id) REFERENCES public.feed(id) ON UPDATE RESTRICT ON DELETE RESTRICT;


--
-- Name: newsletter_feed newsletter_feed_feed_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.newsletter_feed
    ADD CONSTRAINT newsletter_feed_feed_id_fkey FOREIGN KEY (feed_id) REFERENCES public.feed(id) ON UPDATE RESTRICT ON DELETE RESTRICT;


--
-- Name: newsletter_feed_item_status newsletter_feed_item_status_item_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.newsletter_feed_item_status
    ADD CONSTRAINT newsletter_feed_item_status_item_id_fkey FOREIGN KEY (item_id) REFERENCES public.feed_item(id) ON UPDATE RESTRICT ON DELETE RESTRICT;


--
-- Name: newsletter_feed_item_status newsletter_feed_item_status_newsletter_feed_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.newsletter_feed_item_status
    ADD CONSTRAINT newsletter_feed_item_status_newsletter_feed_id_fkey FOREIGN KEY (newsletter_feed_id) REFERENCES public.newsletter_feed(id) ON UPDATE RESTRICT ON DELETE RESTRICT;


--
-- Name: newsletter_feed_item_status newsletter_feed_item_status_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.newsletter_feed_item_status
    ADD CONSTRAINT newsletter_feed_item_status_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.app_user(id) ON UPDATE RESTRICT ON DELETE RESTRICT;


--
-- Name: newsletter_feed newsletter_feed_newsletter_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.newsletter_feed
    ADD CONSTRAINT newsletter_feed_newsletter_id_fkey FOREIGN KEY (newsletter_id) REFERENCES public.newsletter(id) ON UPDATE RESTRICT ON DELETE RESTRICT;


--
-- Name: newsletter_feed newsletter_feed_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.newsletter_feed
    ADD CONSTRAINT newsletter_feed_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.app_user(id) ON UPDATE RESTRICT ON DELETE RESTRICT;


--
-- Name: newsletter newsletter_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.newsletter
    ADD CONSTRAINT newsletter_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.app_user(id) ON UPDATE RESTRICT ON DELETE RESTRICT;


--
-- Name: refresh_token refresh_token_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.refresh_token
    ADD CONSTRAINT refresh_token_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.app_user(id) ON UPDATE RESTRICT ON DELETE RESTRICT;


--
-- PostgreSQL database dump complete
--

\unrestrict 4b4LclgcPQg6rDm8eT15swLJz5mQFDmUqSRzeeOjpgBNyKAnXY52ShEyMdlN5EG


--
-- Dbmate schema migrations
--

INSERT INTO public.schema_migrations (version) VALUES
    ('20260329010424');
