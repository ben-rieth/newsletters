\restrict ZdRU7WdSKgptaBqaz4kO8meeRDkGVQIVEGoxvbfmj1JIXecyeHZoDoftIjIRRxg

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
-- Name: feed_url_source; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.feed_url_source AS ENUM (
    'canonical',
    'user_submitted',
    'in_feed_response',
    'unknown'
);


--
-- Name: filter_field; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.filter_field AS ENUM (
    'title',
    'url'
);


--
-- Name: filter_operator; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.filter_operator AS ENUM (
    'contains',
    'does_not_contain'
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
-- Name: item_state; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.item_state AS ENUM (
    'read',
    'unread'
);


--
-- Name: newsletter_status; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.newsletter_status AS ENUM (
    'active',
    'inactive'
);


--
-- Name: token_purpose; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.token_purpose AS ENUM (
    'email_verify'
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
    email_verified_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    pending_email text DEFAULT ''::text NOT NULL
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
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    html_url text DEFAULT ''::text NOT NULL
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
    source public.feed_url_source NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: issue_item; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.issue_item (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    item_id uuid NOT NULL,
    user_id uuid NOT NULL,
    state public.item_state DEFAULT 'unread'::public.item_state NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    issue_id uuid NOT NULL,
    token uuid DEFAULT gen_random_uuid() NOT NULL
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
    status public.newsletter_status DEFAULT 'active'::public.newsletter_status NOT NULL,
    unsubscribe_token uuid DEFAULT gen_random_uuid() NOT NULL,
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
-- Name: newsletter_feed_filter; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.newsletter_feed_filter (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    newsletter_feed_id uuid NOT NULL,
    user_id uuid NOT NULL,
    field public.filter_field NOT NULL,
    operator public.filter_operator NOT NULL,
    pattern text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: newsletter_issue; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.newsletter_issue (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    newsletter_id uuid NOT NULL,
    user_id uuid NOT NULL,
    sent_at timestamp with time zone DEFAULT now() NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
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
-- Name: verification_token; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.verification_token (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    code text NOT NULL,
    purpose public.token_purpose NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    expires_at timestamp with time zone NOT NULL
);


--
-- Name: white_listed_email; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.white_listed_email (
    email text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
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
-- Name: newsletter_feed_filter newsletter_feed_filter_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.newsletter_feed_filter
    ADD CONSTRAINT newsletter_feed_filter_pkey PRIMARY KEY (id);


--
-- Name: issue_item newsletter_feed_item_status_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.issue_item
    ADD CONSTRAINT newsletter_feed_item_status_pkey PRIMARY KEY (id);


--
-- Name: newsletter_feed newsletter_feed_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.newsletter_feed
    ADD CONSTRAINT newsletter_feed_pkey PRIMARY KEY (id);


--
-- Name: newsletter_issue newsletter_issue_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.newsletter_issue
    ADD CONSTRAINT newsletter_issue_pkey PRIMARY KEY (id);


--
-- Name: newsletter newsletter_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.newsletter
    ADD CONSTRAINT newsletter_pkey PRIMARY KEY (id);


--
-- Name: newsletter newsletter_unsubscribe_token_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.newsletter
    ADD CONSTRAINT newsletter_unsubscribe_token_key UNIQUE (unsubscribe_token);


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
-- Name: verification_token verification_token_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.verification_token
    ADD CONSTRAINT verification_token_pkey PRIMARY KEY (id);


--
-- Name: white_listed_email white_listed_email_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.white_listed_email
    ADD CONSTRAINT white_listed_email_pkey PRIMARY KEY (email);


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
-- Name: issue_item issue_item_issue_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.issue_item
    ADD CONSTRAINT issue_item_issue_id_fkey FOREIGN KEY (issue_id) REFERENCES public.newsletter_issue(id) ON UPDATE RESTRICT ON DELETE RESTRICT;


--
-- Name: newsletter_feed newsletter_feed_feed_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.newsletter_feed
    ADD CONSTRAINT newsletter_feed_feed_id_fkey FOREIGN KEY (feed_id) REFERENCES public.feed(id) ON UPDATE RESTRICT ON DELETE RESTRICT;


--
-- Name: newsletter_feed_filter newsletter_feed_filter_newsletter_feed_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.newsletter_feed_filter
    ADD CONSTRAINT newsletter_feed_filter_newsletter_feed_id_fkey FOREIGN KEY (newsletter_feed_id) REFERENCES public.newsletter_feed(id) ON UPDATE RESTRICT ON DELETE RESTRICT;


--
-- Name: newsletter_feed_filter newsletter_feed_filter_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.newsletter_feed_filter
    ADD CONSTRAINT newsletter_feed_filter_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.app_user(id) ON UPDATE RESTRICT ON DELETE RESTRICT;


--
-- Name: issue_item newsletter_feed_item_status_item_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.issue_item
    ADD CONSTRAINT newsletter_feed_item_status_item_id_fkey FOREIGN KEY (item_id) REFERENCES public.feed_item(id) ON UPDATE RESTRICT ON DELETE RESTRICT;


--
-- Name: issue_item newsletter_feed_item_status_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.issue_item
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
-- Name: newsletter_issue newsletter_issue_newsletter_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.newsletter_issue
    ADD CONSTRAINT newsletter_issue_newsletter_id_fkey FOREIGN KEY (newsletter_id) REFERENCES public.newsletter(id) ON UPDATE RESTRICT ON DELETE RESTRICT;


--
-- Name: newsletter_issue newsletter_issue_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.newsletter_issue
    ADD CONSTRAINT newsletter_issue_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.app_user(id) ON UPDATE RESTRICT ON DELETE RESTRICT;


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
-- Name: verification_token verification_token_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.verification_token
    ADD CONSTRAINT verification_token_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.app_user(id) ON UPDATE RESTRICT ON DELETE RESTRICT;


--
-- PostgreSQL database dump complete
--

\unrestrict ZdRU7WdSKgptaBqaz4kO8meeRDkGVQIVEGoxvbfmj1JIXecyeHZoDoftIjIRRxg


--
-- Dbmate schema migrations
--

INSERT INTO public.schema_migrations (version) VALUES
    ('20260329010424'),
    ('20260421220214'),
    ('20260501141144'),
    ('20260502170343'),
    ('20260509195259');
