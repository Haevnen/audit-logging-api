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
    before_state jsonb,
    after_state jsonb,
    metadata jsonb,
    event_timestamp timestamp with time zone NOT NULL
);


--
-- Name: _direct_view_3; Type: VIEW; Schema: _timescaledb_internal; Owner: -
--

CREATE VIEW _timescaledb_internal._direct_view_3 AS
 SELECT tenant_id,
    public.time_bucket('1 day'::interval, event_timestamp) AS day,
    action,
    severity,
    count(*) AS log_count
   FROM public.logs
  GROUP BY tenant_id, (public.time_bucket('1 day'::interval, event_timestamp)), action, severity;


--
-- Name: _hyper_1_2_chunk; Type: TABLE; Schema: _timescaledb_internal; Owner: -
--

CREATE TABLE _timescaledb_internal._hyper_1_2_chunk (
    CONSTRAINT constraint_3 CHECK (((event_timestamp >= '2025-06-16 00:00:00+00'::timestamp with time zone) AND (event_timestamp < '2025-06-17 00:00:00+00'::timestamp with time zone))),
    CONSTRAINT constraint_4 CHECK (((_timescaledb_functions.get_partition_hash(tenant_id) >= 268435455) AND (_timescaledb_functions.get_partition_hash(tenant_id) < 536870910)))
)
INHERITS (public.logs);


--
-- Name: _hyper_1_5_chunk; Type: TABLE; Schema: _timescaledb_internal; Owner: -
--

CREATE TABLE _timescaledb_internal._hyper_1_5_chunk (
    CONSTRAINT constraint_4 CHECK (((_timescaledb_functions.get_partition_hash(tenant_id) >= 268435455) AND (_timescaledb_functions.get_partition_hash(tenant_id) < 536870910))),
    CONSTRAINT constraint_6 CHECK (((event_timestamp >= '2025-07-15 00:00:00+00'::timestamp with time zone) AND (event_timestamp < '2025-07-16 00:00:00+00'::timestamp with time zone)))
)
INHERITS (public.logs);


--
-- Name: _hyper_1_6_chunk; Type: TABLE; Schema: _timescaledb_internal; Owner: -
--

CREATE TABLE _timescaledb_internal._hyper_1_6_chunk (
    CONSTRAINT constraint_4 CHECK (((_timescaledb_functions.get_partition_hash(tenant_id) >= 268435455) AND (_timescaledb_functions.get_partition_hash(tenant_id) < 536870910))),
    CONSTRAINT constraint_7 CHECK (((event_timestamp >= '2025-07-14 00:00:00+00'::timestamp with time zone) AND (event_timestamp < '2025-07-15 00:00:00+00'::timestamp with time zone)))
)
INHERITS (public.logs);


--
-- Name: _hyper_1_8_chunk; Type: TABLE; Schema: _timescaledb_internal; Owner: -
--

CREATE TABLE _timescaledb_internal._hyper_1_8_chunk (
    CONSTRAINT constraint_3 CHECK (((event_timestamp >= '2025-06-16 00:00:00+00'::timestamp with time zone) AND (event_timestamp < '2025-06-17 00:00:00+00'::timestamp with time zone))),
    CONSTRAINT constraint_9 CHECK ((_timescaledb_functions.get_partition_hash(tenant_id) < 268435455))
)
INHERITS (public.logs);


--
-- Name: _materialized_hypertable_3; Type: TABLE; Schema: _timescaledb_internal; Owner: -
--

CREATE TABLE _timescaledb_internal._materialized_hypertable_3 (
    tenant_id uuid,
    day timestamp with time zone NOT NULL,
    action text,
    severity text,
    log_count bigint
);


--
-- Name: _hyper_3_3_chunk; Type: TABLE; Schema: _timescaledb_internal; Owner: -
--

CREATE TABLE _timescaledb_internal._hyper_3_3_chunk (
    CONSTRAINT constraint_5 CHECK (((day >= '2025-06-11 00:00:00+00'::timestamp with time zone) AND (day < '2025-06-21 00:00:00+00'::timestamp with time zone)))
)
INHERITS (_timescaledb_internal._materialized_hypertable_3);


--
-- Name: _hyper_3_7_chunk; Type: TABLE; Schema: _timescaledb_internal; Owner: -
--

CREATE TABLE _timescaledb_internal._hyper_3_7_chunk (
    CONSTRAINT constraint_8 CHECK (((day >= '2025-07-11 00:00:00+00'::timestamp with time zone) AND (day < '2025-07-21 00:00:00+00'::timestamp with time zone)))
)
INHERITS (_timescaledb_internal._materialized_hypertable_3);


--
-- Name: _partial_view_3; Type: VIEW; Schema: _timescaledb_internal; Owner: -
--

CREATE VIEW _timescaledb_internal._partial_view_3 AS
 SELECT tenant_id,
    public.time_bucket('1 day'::interval, event_timestamp) AS day,
    action,
    severity,
    count(*) AS log_count
   FROM public.logs
  GROUP BY tenant_id, (public.time_bucket('1 day'::interval, event_timestamp)), action, severity;


--
-- Name: compress_hyper_2_4_chunk; Type: TABLE; Schema: _timescaledb_internal; Owner: -
--

CREATE TABLE _timescaledb_internal.compress_hyper_2_4_chunk (
    _ts_meta_count integer,
    tenant_id uuid,
    _ts_meta_v2_bloom1_id _timescaledb_internal.bloom1,
    id _timescaledb_internal.compressed_data,
    user_id _timescaledb_internal.compressed_data,
    session_id _timescaledb_internal.compressed_data,
    action _timescaledb_internal.compressed_data,
    resource _timescaledb_internal.compressed_data,
    resource_id _timescaledb_internal.compressed_data,
    severity _timescaledb_internal.compressed_data,
    ip_address _timescaledb_internal.compressed_data,
    user_agent _timescaledb_internal.compressed_data,
    message _timescaledb_internal.compressed_data,
    before_state _timescaledb_internal.compressed_data,
    after_state _timescaledb_internal.compressed_data,
    metadata _timescaledb_internal.compressed_data,
    _ts_meta_min_1 timestamp with time zone,
    _ts_meta_max_1 timestamp with time zone,
    event_timestamp _timescaledb_internal.compressed_data
)
WITH (toast_tuple_target='128');
ALTER TABLE ONLY _timescaledb_internal.compress_hyper_2_4_chunk ALTER COLUMN _ts_meta_count SET STATISTICS 1000;
ALTER TABLE ONLY _timescaledb_internal.compress_hyper_2_4_chunk ALTER COLUMN tenant_id SET STATISTICS 1000;
ALTER TABLE ONLY _timescaledb_internal.compress_hyper_2_4_chunk ALTER COLUMN _ts_meta_v2_bloom1_id SET STATISTICS 1000;
ALTER TABLE ONLY _timescaledb_internal.compress_hyper_2_4_chunk ALTER COLUMN _ts_meta_v2_bloom1_id SET STORAGE EXTERNAL;
ALTER TABLE ONLY _timescaledb_internal.compress_hyper_2_4_chunk ALTER COLUMN id SET STATISTICS 0;
ALTER TABLE ONLY _timescaledb_internal.compress_hyper_2_4_chunk ALTER COLUMN id SET STORAGE EXTENDED;
ALTER TABLE ONLY _timescaledb_internal.compress_hyper_2_4_chunk ALTER COLUMN user_id SET STATISTICS 0;
ALTER TABLE ONLY _timescaledb_internal.compress_hyper_2_4_chunk ALTER COLUMN user_id SET STORAGE EXTENDED;
ALTER TABLE ONLY _timescaledb_internal.compress_hyper_2_4_chunk ALTER COLUMN session_id SET STATISTICS 0;
ALTER TABLE ONLY _timescaledb_internal.compress_hyper_2_4_chunk ALTER COLUMN session_id SET STORAGE EXTENDED;
ALTER TABLE ONLY _timescaledb_internal.compress_hyper_2_4_chunk ALTER COLUMN action SET STATISTICS 0;
ALTER TABLE ONLY _timescaledb_internal.compress_hyper_2_4_chunk ALTER COLUMN action SET STORAGE EXTENDED;
ALTER TABLE ONLY _timescaledb_internal.compress_hyper_2_4_chunk ALTER COLUMN resource SET STATISTICS 0;
ALTER TABLE ONLY _timescaledb_internal.compress_hyper_2_4_chunk ALTER COLUMN resource SET STORAGE EXTENDED;
ALTER TABLE ONLY _timescaledb_internal.compress_hyper_2_4_chunk ALTER COLUMN resource_id SET STATISTICS 0;
ALTER TABLE ONLY _timescaledb_internal.compress_hyper_2_4_chunk ALTER COLUMN resource_id SET STORAGE EXTENDED;
ALTER TABLE ONLY _timescaledb_internal.compress_hyper_2_4_chunk ALTER COLUMN severity SET STATISTICS 0;
ALTER TABLE ONLY _timescaledb_internal.compress_hyper_2_4_chunk ALTER COLUMN severity SET STORAGE EXTENDED;
ALTER TABLE ONLY _timescaledb_internal.compress_hyper_2_4_chunk ALTER COLUMN ip_address SET STATISTICS 0;
ALTER TABLE ONLY _timescaledb_internal.compress_hyper_2_4_chunk ALTER COLUMN ip_address SET STORAGE EXTENDED;
ALTER TABLE ONLY _timescaledb_internal.compress_hyper_2_4_chunk ALTER COLUMN user_agent SET STATISTICS 0;
ALTER TABLE ONLY _timescaledb_internal.compress_hyper_2_4_chunk ALTER COLUMN user_agent SET STORAGE EXTENDED;
ALTER TABLE ONLY _timescaledb_internal.compress_hyper_2_4_chunk ALTER COLUMN message SET STATISTICS 0;
ALTER TABLE ONLY _timescaledb_internal.compress_hyper_2_4_chunk ALTER COLUMN message SET STORAGE EXTENDED;
ALTER TABLE ONLY _timescaledb_internal.compress_hyper_2_4_chunk ALTER COLUMN before_state SET STATISTICS 0;
ALTER TABLE ONLY _timescaledb_internal.compress_hyper_2_4_chunk ALTER COLUMN before_state SET STORAGE EXTENDED;
ALTER TABLE ONLY _timescaledb_internal.compress_hyper_2_4_chunk ALTER COLUMN after_state SET STATISTICS 0;
ALTER TABLE ONLY _timescaledb_internal.compress_hyper_2_4_chunk ALTER COLUMN after_state SET STORAGE EXTENDED;
ALTER TABLE ONLY _timescaledb_internal.compress_hyper_2_4_chunk ALTER COLUMN metadata SET STATISTICS 0;
ALTER TABLE ONLY _timescaledb_internal.compress_hyper_2_4_chunk ALTER COLUMN metadata SET STORAGE EXTENDED;
ALTER TABLE ONLY _timescaledb_internal.compress_hyper_2_4_chunk ALTER COLUMN _ts_meta_min_1 SET STATISTICS 1000;
ALTER TABLE ONLY _timescaledb_internal.compress_hyper_2_4_chunk ALTER COLUMN _ts_meta_max_1 SET STATISTICS 1000;
ALTER TABLE ONLY _timescaledb_internal.compress_hyper_2_4_chunk ALTER COLUMN event_timestamp SET STATISTICS 0;


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
-- Name: log_stats_daily; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.log_stats_daily AS
 SELECT tenant_id,
    day,
    action,
    severity,
    log_count
   FROM _timescaledb_internal._materialized_hypertable_3;


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
-- Name: _hyper_1_2_chunk id; Type: DEFAULT; Schema: _timescaledb_internal; Owner: -
--

ALTER TABLE ONLY _timescaledb_internal._hyper_1_2_chunk ALTER COLUMN id SET DEFAULT gen_random_uuid();


--
-- Name: _hyper_1_2_chunk severity; Type: DEFAULT; Schema: _timescaledb_internal; Owner: -
--

ALTER TABLE ONLY _timescaledb_internal._hyper_1_2_chunk ALTER COLUMN severity SET DEFAULT 'INFO'::text;


--
-- Name: _hyper_1_5_chunk id; Type: DEFAULT; Schema: _timescaledb_internal; Owner: -
--

ALTER TABLE ONLY _timescaledb_internal._hyper_1_5_chunk ALTER COLUMN id SET DEFAULT gen_random_uuid();


--
-- Name: _hyper_1_5_chunk severity; Type: DEFAULT; Schema: _timescaledb_internal; Owner: -
--

ALTER TABLE ONLY _timescaledb_internal._hyper_1_5_chunk ALTER COLUMN severity SET DEFAULT 'INFO'::text;


--
-- Name: _hyper_1_6_chunk id; Type: DEFAULT; Schema: _timescaledb_internal; Owner: -
--

ALTER TABLE ONLY _timescaledb_internal._hyper_1_6_chunk ALTER COLUMN id SET DEFAULT gen_random_uuid();


--
-- Name: _hyper_1_6_chunk severity; Type: DEFAULT; Schema: _timescaledb_internal; Owner: -
--

ALTER TABLE ONLY _timescaledb_internal._hyper_1_6_chunk ALTER COLUMN severity SET DEFAULT 'INFO'::text;


--
-- Name: _hyper_1_8_chunk id; Type: DEFAULT; Schema: _timescaledb_internal; Owner: -
--

ALTER TABLE ONLY _timescaledb_internal._hyper_1_8_chunk ALTER COLUMN id SET DEFAULT gen_random_uuid();


--
-- Name: _hyper_1_8_chunk severity; Type: DEFAULT; Schema: _timescaledb_internal; Owner: -
--

ALTER TABLE ONLY _timescaledb_internal._hyper_1_8_chunk ALTER COLUMN severity SET DEFAULT 'INFO'::text;


--
-- Name: _hyper_1_2_chunk 2_3_logs_pkey; Type: CONSTRAINT; Schema: _timescaledb_internal; Owner: -
--

ALTER TABLE ONLY _timescaledb_internal._hyper_1_2_chunk
    ADD CONSTRAINT "2_3_logs_pkey" PRIMARY KEY (tenant_id, event_timestamp, id);


--
-- Name: _hyper_1_5_chunk 5_5_logs_pkey; Type: CONSTRAINT; Schema: _timescaledb_internal; Owner: -
--

ALTER TABLE ONLY _timescaledb_internal._hyper_1_5_chunk
    ADD CONSTRAINT "5_5_logs_pkey" PRIMARY KEY (tenant_id, event_timestamp, id);


--
-- Name: _hyper_1_6_chunk 6_7_logs_pkey; Type: CONSTRAINT; Schema: _timescaledb_internal; Owner: -
--

ALTER TABLE ONLY _timescaledb_internal._hyper_1_6_chunk
    ADD CONSTRAINT "6_7_logs_pkey" PRIMARY KEY (tenant_id, event_timestamp, id);


--
-- Name: _hyper_1_8_chunk 8_9_logs_pkey; Type: CONSTRAINT; Schema: _timescaledb_internal; Owner: -
--

ALTER TABLE ONLY _timescaledb_internal._hyper_1_8_chunk
    ADD CONSTRAINT "8_9_logs_pkey" PRIMARY KEY (tenant_id, event_timestamp, id);


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
-- Name: _hyper_1_2_chunk_logs_event_timestamp_idx; Type: INDEX; Schema: _timescaledb_internal; Owner: -
--

CREATE INDEX _hyper_1_2_chunk_logs_event_timestamp_idx ON _timescaledb_internal._hyper_1_2_chunk USING btree (event_timestamp DESC);


--
-- Name: _hyper_1_2_chunk_logs_tenant_id_event_timestamp_idx; Type: INDEX; Schema: _timescaledb_internal; Owner: -
--

CREATE INDEX _hyper_1_2_chunk_logs_tenant_id_event_timestamp_idx ON _timescaledb_internal._hyper_1_2_chunk USING btree (tenant_id, event_timestamp DESC);


--
-- Name: _hyper_1_5_chunk_logs_event_timestamp_idx; Type: INDEX; Schema: _timescaledb_internal; Owner: -
--

CREATE INDEX _hyper_1_5_chunk_logs_event_timestamp_idx ON _timescaledb_internal._hyper_1_5_chunk USING btree (event_timestamp DESC);


--
-- Name: _hyper_1_5_chunk_logs_tenant_id_event_timestamp_idx; Type: INDEX; Schema: _timescaledb_internal; Owner: -
--

CREATE INDEX _hyper_1_5_chunk_logs_tenant_id_event_timestamp_idx ON _timescaledb_internal._hyper_1_5_chunk USING btree (tenant_id, event_timestamp DESC);


--
-- Name: _hyper_1_6_chunk_logs_event_timestamp_idx; Type: INDEX; Schema: _timescaledb_internal; Owner: -
--

CREATE INDEX _hyper_1_6_chunk_logs_event_timestamp_idx ON _timescaledb_internal._hyper_1_6_chunk USING btree (event_timestamp DESC);


--
-- Name: _hyper_1_6_chunk_logs_tenant_id_event_timestamp_idx; Type: INDEX; Schema: _timescaledb_internal; Owner: -
--

CREATE INDEX _hyper_1_6_chunk_logs_tenant_id_event_timestamp_idx ON _timescaledb_internal._hyper_1_6_chunk USING btree (tenant_id, event_timestamp DESC);


--
-- Name: _hyper_1_8_chunk_logs_event_timestamp_idx; Type: INDEX; Schema: _timescaledb_internal; Owner: -
--

CREATE INDEX _hyper_1_8_chunk_logs_event_timestamp_idx ON _timescaledb_internal._hyper_1_8_chunk USING btree (event_timestamp DESC);


--
-- Name: _hyper_1_8_chunk_logs_tenant_id_event_timestamp_idx; Type: INDEX; Schema: _timescaledb_internal; Owner: -
--

CREATE INDEX _hyper_1_8_chunk_logs_tenant_id_event_timestamp_idx ON _timescaledb_internal._hyper_1_8_chunk USING btree (tenant_id, event_timestamp DESC);


--
-- Name: _hyper_3_3_chunk__materialized_hypertable_3_action_day_idx; Type: INDEX; Schema: _timescaledb_internal; Owner: -
--

CREATE INDEX _hyper_3_3_chunk__materialized_hypertable_3_action_day_idx ON _timescaledb_internal._hyper_3_3_chunk USING btree (action, day DESC);


--
-- Name: _hyper_3_3_chunk__materialized_hypertable_3_day_idx; Type: INDEX; Schema: _timescaledb_internal; Owner: -
--

CREATE INDEX _hyper_3_3_chunk__materialized_hypertable_3_day_idx ON _timescaledb_internal._hyper_3_3_chunk USING btree (day DESC);


--
-- Name: _hyper_3_3_chunk__materialized_hypertable_3_severity_day_idx; Type: INDEX; Schema: _timescaledb_internal; Owner: -
--

CREATE INDEX _hyper_3_3_chunk__materialized_hypertable_3_severity_day_idx ON _timescaledb_internal._hyper_3_3_chunk USING btree (severity, day DESC);


--
-- Name: _hyper_3_3_chunk__materialized_hypertable_3_tenant_id_day_idx; Type: INDEX; Schema: _timescaledb_internal; Owner: -
--

CREATE INDEX _hyper_3_3_chunk__materialized_hypertable_3_tenant_id_day_idx ON _timescaledb_internal._hyper_3_3_chunk USING btree (tenant_id, day DESC);


--
-- Name: _hyper_3_3_chunk_idx_log_stats_daily_tenant_day; Type: INDEX; Schema: _timescaledb_internal; Owner: -
--

CREATE INDEX _hyper_3_3_chunk_idx_log_stats_daily_tenant_day ON _timescaledb_internal._hyper_3_3_chunk USING btree (tenant_id, day);


--
-- Name: _hyper_3_7_chunk__materialized_hypertable_3_action_day_idx; Type: INDEX; Schema: _timescaledb_internal; Owner: -
--

CREATE INDEX _hyper_3_7_chunk__materialized_hypertable_3_action_day_idx ON _timescaledb_internal._hyper_3_7_chunk USING btree (action, day DESC);


--
-- Name: _hyper_3_7_chunk__materialized_hypertable_3_day_idx; Type: INDEX; Schema: _timescaledb_internal; Owner: -
--

CREATE INDEX _hyper_3_7_chunk__materialized_hypertable_3_day_idx ON _timescaledb_internal._hyper_3_7_chunk USING btree (day DESC);


--
-- Name: _hyper_3_7_chunk__materialized_hypertable_3_severity_day_idx; Type: INDEX; Schema: _timescaledb_internal; Owner: -
--

CREATE INDEX _hyper_3_7_chunk__materialized_hypertable_3_severity_day_idx ON _timescaledb_internal._hyper_3_7_chunk USING btree (severity, day DESC);


--
-- Name: _hyper_3_7_chunk__materialized_hypertable_3_tenant_id_day_idx; Type: INDEX; Schema: _timescaledb_internal; Owner: -
--

CREATE INDEX _hyper_3_7_chunk__materialized_hypertable_3_tenant_id_day_idx ON _timescaledb_internal._hyper_3_7_chunk USING btree (tenant_id, day DESC);


--
-- Name: _hyper_3_7_chunk_idx_log_stats_daily_tenant_day; Type: INDEX; Schema: _timescaledb_internal; Owner: -
--

CREATE INDEX _hyper_3_7_chunk_idx_log_stats_daily_tenant_day ON _timescaledb_internal._hyper_3_7_chunk USING btree (tenant_id, day);


--
-- Name: _materialized_hypertable_3_action_day_idx; Type: INDEX; Schema: _timescaledb_internal; Owner: -
--

CREATE INDEX _materialized_hypertable_3_action_day_idx ON _timescaledb_internal._materialized_hypertable_3 USING btree (action, day DESC);


--
-- Name: _materialized_hypertable_3_day_idx; Type: INDEX; Schema: _timescaledb_internal; Owner: -
--

CREATE INDEX _materialized_hypertable_3_day_idx ON _timescaledb_internal._materialized_hypertable_3 USING btree (day DESC);


--
-- Name: _materialized_hypertable_3_severity_day_idx; Type: INDEX; Schema: _timescaledb_internal; Owner: -
--

CREATE INDEX _materialized_hypertable_3_severity_day_idx ON _timescaledb_internal._materialized_hypertable_3 USING btree (severity, day DESC);


--
-- Name: _materialized_hypertable_3_tenant_id_day_idx; Type: INDEX; Schema: _timescaledb_internal; Owner: -
--

CREATE INDEX _materialized_hypertable_3_tenant_id_day_idx ON _timescaledb_internal._materialized_hypertable_3 USING btree (tenant_id, day DESC);


--
-- Name: compress_hyper_2_4_chunk_tenant_id__ts_meta_min_1__ts_meta__idx; Type: INDEX; Schema: _timescaledb_internal; Owner: -
--

CREATE INDEX compress_hyper_2_4_chunk_tenant_id__ts_meta_min_1__ts_meta__idx ON _timescaledb_internal.compress_hyper_2_4_chunk USING btree (tenant_id, _ts_meta_min_1 DESC, _ts_meta_max_1 DESC);


--
-- Name: idx_log_stats_daily_tenant_day; Type: INDEX; Schema: _timescaledb_internal; Owner: -
--

CREATE INDEX idx_log_stats_daily_tenant_day ON _timescaledb_internal._materialized_hypertable_3 USING btree (tenant_id, day);


--
-- Name: logs_event_timestamp_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX logs_event_timestamp_idx ON public.logs USING btree (event_timestamp DESC);


--
-- Name: logs_tenant_id_event_timestamp_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX logs_tenant_id_event_timestamp_idx ON public.logs USING btree (tenant_id, event_timestamp DESC);


--
-- Name: _hyper_1_2_chunk ts_cagg_invalidation_trigger; Type: TRIGGER; Schema: _timescaledb_internal; Owner: -
--

CREATE TRIGGER ts_cagg_invalidation_trigger AFTER INSERT OR DELETE OR UPDATE ON _timescaledb_internal._hyper_1_2_chunk FOR EACH ROW EXECUTE FUNCTION _timescaledb_functions.continuous_agg_invalidation_trigger('1');


--
-- Name: _hyper_1_5_chunk ts_cagg_invalidation_trigger; Type: TRIGGER; Schema: _timescaledb_internal; Owner: -
--

CREATE TRIGGER ts_cagg_invalidation_trigger AFTER INSERT OR DELETE OR UPDATE ON _timescaledb_internal._hyper_1_5_chunk FOR EACH ROW EXECUTE FUNCTION _timescaledb_functions.continuous_agg_invalidation_trigger('1');


--
-- Name: _hyper_1_6_chunk ts_cagg_invalidation_trigger; Type: TRIGGER; Schema: _timescaledb_internal; Owner: -
--

CREATE TRIGGER ts_cagg_invalidation_trigger AFTER INSERT OR DELETE OR UPDATE ON _timescaledb_internal._hyper_1_6_chunk FOR EACH ROW EXECUTE FUNCTION _timescaledb_functions.continuous_agg_invalidation_trigger('1');


--
-- Name: _hyper_1_8_chunk ts_cagg_invalidation_trigger; Type: TRIGGER; Schema: _timescaledb_internal; Owner: -
--

CREATE TRIGGER ts_cagg_invalidation_trigger AFTER INSERT OR DELETE OR UPDATE ON _timescaledb_internal._hyper_1_8_chunk FOR EACH ROW EXECUTE FUNCTION _timescaledb_functions.continuous_agg_invalidation_trigger('1');


--
-- Name: _compressed_hypertable_2 ts_insert_blocker; Type: TRIGGER; Schema: _timescaledb_internal; Owner: -
--

CREATE TRIGGER ts_insert_blocker BEFORE INSERT ON _timescaledb_internal._compressed_hypertable_2 FOR EACH ROW EXECUTE FUNCTION _timescaledb_functions.insert_blocker();


--
-- Name: _materialized_hypertable_3 ts_insert_blocker; Type: TRIGGER; Schema: _timescaledb_internal; Owner: -
--

CREATE TRIGGER ts_insert_blocker BEFORE INSERT ON _timescaledb_internal._materialized_hypertable_3 FOR EACH ROW EXECUTE FUNCTION _timescaledb_functions.insert_blocker();


--
-- Name: logs ts_cagg_invalidation_trigger; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER ts_cagg_invalidation_trigger AFTER INSERT OR DELETE OR UPDATE ON public.logs FOR EACH ROW EXECUTE FUNCTION _timescaledb_functions.continuous_agg_invalidation_trigger('1');


--
-- Name: logs ts_insert_blocker; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER ts_insert_blocker BEFORE INSERT ON public.logs FOR EACH ROW EXECUTE FUNCTION _timescaledb_functions.insert_blocker();


--
-- Name: _hyper_1_2_chunk 2_4_logs_tenant_id_fkey; Type: FK CONSTRAINT; Schema: _timescaledb_internal; Owner: -
--

ALTER TABLE ONLY _timescaledb_internal._hyper_1_2_chunk
    ADD CONSTRAINT "2_4_logs_tenant_id_fkey" FOREIGN KEY (tenant_id) REFERENCES public.tenants(id) ON DELETE CASCADE;


--
-- Name: _hyper_1_5_chunk 5_6_logs_tenant_id_fkey; Type: FK CONSTRAINT; Schema: _timescaledb_internal; Owner: -
--

ALTER TABLE ONLY _timescaledb_internal._hyper_1_5_chunk
    ADD CONSTRAINT "5_6_logs_tenant_id_fkey" FOREIGN KEY (tenant_id) REFERENCES public.tenants(id) ON DELETE CASCADE;


--
-- Name: _hyper_1_6_chunk 6_8_logs_tenant_id_fkey; Type: FK CONSTRAINT; Schema: _timescaledb_internal; Owner: -
--

ALTER TABLE ONLY _timescaledb_internal._hyper_1_6_chunk
    ADD CONSTRAINT "6_8_logs_tenant_id_fkey" FOREIGN KEY (tenant_id) REFERENCES public.tenants(id) ON DELETE CASCADE;


--
-- Name: _hyper_1_8_chunk 8_10_logs_tenant_id_fkey; Type: FK CONSTRAINT; Schema: _timescaledb_internal; Owner: -
--

ALTER TABLE ONLY _timescaledb_internal._hyper_1_8_chunk
    ADD CONSTRAINT "8_10_logs_tenant_id_fkey" FOREIGN KEY (tenant_id) REFERENCES public.tenants(id) ON DELETE CASCADE;


--
-- Name: logs logs_tenant_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.logs
    ADD CONSTRAINT logs_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES public.tenants(id) ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--

