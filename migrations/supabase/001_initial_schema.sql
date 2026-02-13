-- Supabase Migration: Initial Schema for Invento Service
-- This migration creates the database schema for the application

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================================================
-- ROLES & PERMISSIONS TABLES (RBAC)
-- ============================================================================

CREATE TABLE IF NOT EXISTS roles (
    id SERIAL PRIMARY KEY,
    nama_role VARCHAR(100) NOT NULL UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS permissions (
    id SERIAL PRIMARY KEY,
    resource VARCHAR(100) NOT NULL,
    action VARCHAR(50) NOT NULL,
    label VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(resource, action)
);

CREATE TABLE IF NOT EXISTS role_permissions (
    id SERIAL PRIMARY KEY,
    role_id INTEGER REFERENCES roles(id) ON DELETE CASCADE,
    permission_id INTEGER REFERENCES permissions(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(role_id, permission_id)
);

-- Casbin rule table for RBAC policies
CREATE TABLE IF NOT EXISTS casbin_rule (
    id SERIAL PRIMARY KEY,
    ptype VARCHAR(255),
    v0 VARCHAR(255),
    v1 VARCHAR(255),
    v2 VARCHAR(255),
    v3 VARCHAR(255),
    v4 VARCHAR(255),
    v5 VARCHAR(255)
);

-- ============================================================================
-- USER PROFILES (links to Supabase auth.users)
-- ============================================================================

CREATE TABLE IF NOT EXISTS user_profiles (
    id UUID REFERENCES auth.users(id) ON DELETE CASCADE PRIMARY KEY,
    email TEXT NOT NULL,
    name TEXT NOT NULL,
    jenis_kelamin TEXT,
    foto_profil TEXT,
    role_id INTEGER REFERENCES roles(id) ON DELETE SET NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(email)
);

-- Indexes for user_profiles
CREATE INDEX IF NOT EXISTS idx_user_profiles_email ON user_profiles(email);
CREATE INDEX IF NOT EXISTS idx_user_profiles_role_id ON user_profiles(role_id);
CREATE INDEX IF NOT EXISTS idx_user_profiles_is_active ON user_profiles(is_active);

-- Enable Row Level Security
ALTER TABLE user_profiles ENABLE ROW LEVEL SECURITY;

-- RLS Policies for user_profiles
CREATE POLICY "Users can view own profile" ON user_profiles
    FOR SELECT USING (auth.uid() = id);

CREATE POLICY "Service role can manage all profiles" ON user_profiles
    FOR ALL USING (auth.jwt() ->> 'role' = 'service_role');

CREATE POLICY "Authenticated users can view profiles" ON user_profiles
    FOR SELECT USING (auth.role() = 'authenticated');

-- ============================================================================
-- PROJECTS TABLE
-- ============================================================================

CREATE TABLE IF NOT EXISTS projects (
    id SERIAL PRIMARY KEY,
    user_id UUID REFERENCES auth.users(id) ON DELETE CASCADE NOT NULL,
    nama_project VARCHAR(255) NOT NULL,
    kategori VARCHAR(50) NOT NULL DEFAULT '',
    semester INTEGER NOT NULL DEFAULT 1,
    ukuran VARCHAR(50) NOT NULL DEFAULT '',
    path_file VARCHAR(500) NOT NULL DEFAULT '',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_projects_user_id ON projects(user_id);
CREATE INDEX IF NOT EXISTS idx_projects_status ON projects(status);

ALTER TABLE projects ENABLE ROW LEVEL SECURITY;

CREATE POLICY "Users can view own projects" ON projects
    FOR SELECT USING (auth.uid() = user_id);

CREATE POLICY "Users can create own projects" ON projects
    FOR INSERT WITH CHECK (auth.uid() = user_id);

CREATE POLICY "Users can update own projects" ON projects
    FOR UPDATE USING (auth.uid() = user_id);

CREATE POLICY "Users can delete own projects" ON projects
    FOR DELETE USING (auth.uid() = user_id);

CREATE POLICY "Service role full access projects" ON projects
    FOR ALL USING (auth.jwt() ->> 'role' = 'service_role');

-- ============================================================================
-- MODULS TABLE
-- ============================================================================

CREATE TABLE IF NOT EXISTS moduls (
    id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    user_id UUID REFERENCES auth.users(id) ON DELETE CASCADE NOT NULL,
    judul VARCHAR(255) NOT NULL,
    deskripsi TEXT,
    file_path VARCHAR(500),
    file_name VARCHAR(255),
    file_size BIGINT,
    mime_type VARCHAR(100),
    status VARCHAR(50) DEFAULT 'pending',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_moduls_user_id ON moduls(user_id);
CREATE INDEX IF NOT EXISTS idx_moduls_status ON moduls(status);

ALTER TABLE moduls ENABLE ROW LEVEL SECURITY;

CREATE POLICY "Users can view own moduls" ON moduls
    FOR SELECT USING (auth.uid() = user_id);

CREATE POLICY "Users can create own moduls" ON moduls
    FOR INSERT WITH CHECK (auth.uid() = user_id);

CREATE POLICY "Users can update own moduls" ON moduls
    FOR UPDATE USING (auth.uid() = user_id);

CREATE POLICY "Users can delete own moduls" ON moduls
    FOR DELETE USING (auth.uid() = user_id);

CREATE POLICY "Service role full access moduls" ON moduls
    FOR ALL USING (auth.jwt() ->> 'role' = 'service_role');

-- ============================================================================
-- TUS UPLOADS TABLE (for project file uploads)
-- ============================================================================

CREATE TABLE IF NOT EXISTS tus_uploads (
    id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    user_id UUID REFERENCES auth.users(id) ON DELETE CASCADE NOT NULL,
    project_id INTEGER REFERENCES projects(id) ON DELETE SET NULL,
    upload_url VARCHAR(500) UNIQUE NOT NULL,
    upload_metadata JSONB,
    upload_type VARCHAR(20) NOT NULL DEFAULT 'project_create',
    file_size BIGINT,
    current_offset BIGINT DEFAULT 0,
    file_path VARCHAR(500),
    status VARCHAR(50) DEFAULT 'pending',
    progress REAL DEFAULT 0,
    completed_at TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_tus_uploads_user_id ON tus_uploads(user_id);
CREATE INDEX IF NOT EXISTS idx_tus_uploads_project_id ON tus_uploads(project_id);
CREATE INDEX IF NOT EXISTS idx_tus_uploads_status ON tus_uploads(status);

ALTER TABLE tus_uploads ENABLE ROW LEVEL SECURITY;

CREATE POLICY "Users can view own tus uploads" ON tus_uploads
    FOR SELECT USING (auth.uid() = user_id);

CREATE POLICY "Users can create own tus uploads" ON tus_uploads
    FOR INSERT WITH CHECK (auth.uid() = user_id);

CREATE POLICY "Users can update own tus uploads" ON tus_uploads
    FOR UPDATE USING (auth.uid() = user_id);

CREATE POLICY "Service role full access tus uploads" ON tus_uploads
    FOR ALL USING (auth.jwt() ->> 'role' = 'service_role');

-- ============================================================================
-- TUS MODUL UPLOADS TABLE (for modul file uploads)
-- ============================================================================

CREATE TABLE IF NOT EXISTS tus_modul_uploads (
    id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    user_id UUID REFERENCES auth.users(id) ON DELETE CASCADE NOT NULL,
    modul_id UUID REFERENCES moduls(id) ON DELETE SET NULL,
    upload_url VARCHAR(500) UNIQUE NOT NULL,
    upload_metadata JSONB,
    upload_type VARCHAR(20) NOT NULL DEFAULT 'modul_create',
    file_size BIGINT,
    current_offset BIGINT DEFAULT 0,
    file_path VARCHAR(500),
    status VARCHAR(50) DEFAULT 'pending',
    progress REAL DEFAULT 0,
    completed_at TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_tus_modul_uploads_user_id ON tus_modul_uploads(user_id);
CREATE INDEX IF NOT EXISTS idx_tus_modul_uploads_modul_id ON tus_modul_uploads(modul_id);
CREATE INDEX IF NOT EXISTS idx_tus_modul_uploads_status ON tus_modul_uploads(status);

ALTER TABLE tus_modul_uploads ENABLE ROW LEVEL SECURITY;

CREATE POLICY "Users can view own tus modul uploads" ON tus_modul_uploads
    FOR SELECT USING (auth.uid() = user_id);

CREATE POLICY "Users can create own tus modul uploads" ON tus_modul_uploads
    FOR INSERT WITH CHECK (auth.uid() = user_id);

CREATE POLICY "Users can update own tus modul uploads" ON tus_modul_uploads
    FOR UPDATE USING (auth.uid() = user_id);

CREATE POLICY "Service role full access tus modul uploads" ON tus_modul_uploads
    FOR ALL USING (auth.jwt() ->> 'role' = 'service_role');

-- ============================================================================
-- TRIGGER: Auto-create user profile on signup
-- ============================================================================

CREATE OR REPLACE FUNCTION public.handle_new_user()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO public.user_profiles (id, email, name, is_active)
    VALUES (
        NEW.id,
        NEW.email,
        COALESCE(NEW.raw_user_meta_data->>'name', 'User'),
        true
    );
    RETURN NEW;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Drop trigger if exists and create new one
DROP TRIGGER IF EXISTS on_auth_user_created ON auth.users;
CREATE TRIGGER on_auth_user_created
    AFTER INSERT ON auth.users
    FOR EACH ROW EXECUTE FUNCTION public.handle_new_user();

-- ============================================================================
-- FUNCTION: Update updated_at timestamp
-- ============================================================================

CREATE OR REPLACE FUNCTION public.update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Add updated_at triggers
CREATE TRIGGER update_user_profiles_updated_at BEFORE UPDATE ON user_profiles
    FOR EACH ROW EXECUTE FUNCTION public.update_updated_at();

CREATE TRIGGER update_projects_updated_at BEFORE UPDATE ON projects
    FOR EACH ROW EXECUTE FUNCTION public.update_updated_at();

CREATE TRIGGER update_moduls_updated_at BEFORE UPDATE ON moduls
    FOR EACH ROW EXECUTE FUNCTION public.update_updated_at();

CREATE TRIGGER update_tus_uploads_updated_at BEFORE UPDATE ON tus_uploads
    FOR EACH ROW EXECUTE FUNCTION public.update_updated_at();

CREATE TRIGGER update_tus_modul_uploads_updated_at BEFORE UPDATE ON tus_modul_uploads
    FOR EACH ROW EXECUTE FUNCTION public.update_updated_at();
