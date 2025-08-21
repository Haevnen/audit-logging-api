--
-- PostgreSQL database dump
--

-- Dumped from database version 16.9
-- Dumped by pg_dump version 16.9

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
-- Name: timescaledb; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS timescaledb WITH SCHEMA public;


--
-- Name: EXTENSION timescaledb; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON EXTENSION timescaledb IS 'Enables scalable inserts and complex queries for time-series data (Community Edition)';


--
-- Name: async_task_status; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.async_task_status AS ENUM (
    'pending',
    'running',
    'succeeded',
    'failed'
);


--
-- Name: async_task_type; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.async_task_type AS ENUM (
    'log_cleanup',
    'archive',
    'export',
    'reindex'
);


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: _compressed_hypertable_2; Type: TABLE; Schema: _timescaledb_internal; Owner: -
--

CREATE TABLE _timescaledb_internal._compressed_hypertable_2 (
);


--
-- Name: async_tasks; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.async_tasks (
    task_id uuid DEFAULT gen_random_uuid() NOT NULL,
    status public.async_task_status NOT NULL,
    task_type public.async_task_type NOT NULL,
    payload jsonb,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    tenant_uid text,
    user_id text NOT NULL,
    error_msg text
);


--
-- Name: logs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.logs (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    tenant_id uuid NOT NULL,
    user_id text NOT NULL,
    session_id text,
    action text NOT NULL,
    resource text,
    resource_id text,
    severity text DEFAULT 'INFO'::text NOT NULL,
    ip_address inet,
    user_agent text,
    message text,
    before jsonb,
    after jsonb,
    metadata jsonb,
    event_timestamp timestamp with time zone NOT NULL
);


--
-- Name: tenants; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.tenants (
    id uuid NOT NULL,
    name text NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


--
-- Name: async_tasks async_tasks_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.async_tasks
    ADD CONSTRAINT async_tasks_pkey PRIMARY KEY (task_id);


--
-- Name: logs logs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.logs
    ADD CONSTRAINT logs_pkey PRIMARY KEY (tenant_id, event_timestamp, id);


--
-- Name: tenants tenants_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tenants
    ADD CONSTRAINT tenants_pkey PRIMARY KEY (id);


--
-- Name: logs_event_timestamp_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX logs_event_timestamp_idx ON public.logs USING btree (event_timestamp DESC);


--
-- Name: logs_tenant_id_event_timestamp_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX logs_tenant_id_event_timestamp_idx ON public.logs USING btree (tenant_id, event_timestamp DESC);


--
-- Name: _compressed_hypertable_2 ts_insert_blocker; Type: TRIGGER; Schema: _timescaledb_internal; Owner: -
--

CREATE TRIGGER ts_insert_blocker BEFORE INSERT ON _timescaledb_internal._compressed_hypertable_2 FOR EACH ROW EXECUTE FUNCTION _timescaledb_functions.insert_blocker();


--
-- Name: logs ts_insert_blocker; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER ts_insert_blocker BEFORE INSERT ON public.logs FOR EACH ROW EXECUTE FUNCTION _timescaledb_functions.insert_blocker();


--
-- Name: logs logs_tenant_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.logs
    ADD CONSTRAINT logs_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES public.tenants(id) ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--

