--
-- PostgreSQL database dump
--

-- Dumped from database version 13.4
-- Dumped by pg_dump version 13.4

-- Started on 2022-10-11 13:01:27 +04

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
-- TOC entry 200 (class 1259 OID 66307)
-- Name: order; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public."order" (
    uid character varying(32) NOT NULL,
    track_number character varying(32) NOT NULL,
    record jsonb NOT NULL
);


ALTER TABLE public."order" OWNER TO postgres;

--
-- TOC entry 3858 (class 2606 OID 66319)
-- Name: order order_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public."order"
    ADD CONSTRAINT order_pkey PRIMARY KEY (uid, track_number);


--
-- TOC entry 3859 (class 1259 OID 66317)
-- Name: order_track_number_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX order_track_number_idx ON public."order" USING btree (track_number);


--
-- TOC entry 3860 (class 1259 OID 66316)
-- Name: order_uid_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX order_uid_idx ON public."order" USING btree (uid);


-- Completed on 2022-10-11 13:01:27 +04

--
-- PostgreSQL database dump complete
--

