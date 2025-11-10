-- This schema is for context only and is not meant to be run.
-- Table order and constraints may not be valid for execution.

CREATE TABLE public.users (
  id uuid NOT NULL DEFAULT gen_random_uuid(),
  username text UNIQUE,
  email text NOT NULL UNIQUE,
  password_hash text,
  credits bigint DEFAULT 10,
  stripe_customer_id text UNIQUE,
  refresh_token text UNIQUE,
  password_reset_token text,
  password_reset_expires timestamp with time zone,
  is_email_verified boolean DEFAULT false,
  email_verification_token text,
  email_verification_expires timestamp with time zone,
  created_at timestamp with time zone,
  updated_at timestamp with time zone,
  deleted_at timestamp with time zone,
  CONSTRAINT users_pkey PRIMARY KEY (id)
);

CREATE TABLE public.uploaded_files (
  id uuid NOT NULL DEFAULT gen_random_uuid(),
  created_at timestamp with time zone NOT NULL DEFAULT now(),
  updated_at timestamp without time zone DEFAULT now(),
  user_id uuid NOT NULL,
  display_name text NOT NULL DEFAULT ''::text,
  status text NOT NULL DEFAULT 'queue'::text,
  uploaded boolean NOT NULL,
  CONSTRAINT uploaded_files_pkey PRIMARY KEY (id),
  CONSTRAINT uploaded_files_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE
);

CREATE TABLE public.clips (
  id uuid NOT NULL DEFAULT gen_random_uuid(),
  created_at timestamp with time zone NOT NULL DEFAULT now(),
  updated_at timestamp without time zone,
  user_id uuid NOT NULL,
  uploaded_file_id uuid NOT NULL,
  CONSTRAINT clips_pkey PRIMARY KEY (id),
  CONSTRAINT clips_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE,
  CONSTRAINT clips_uploaded_file_id_fkey FOREIGN KEY (uploaded_file_id) REFERENCES public.uploaded_files(id) ON DELETE CASCADE
);
