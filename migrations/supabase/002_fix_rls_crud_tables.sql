-- ================================================
-- RLS Policy Migration: CRUD Tables
-- Fix: InitPlan caching (RLS-01), service role scoping (RLS-02), user role scoping (RLS-03)
-- Tables: projects, moduls, tus_uploads, tus_modul_uploads
-- Applied via Supabase MCP apply_migration
-- ================================================

-- ---- PROJECTS ----

ALTER POLICY "Users can view own projects" ON projects
  TO authenticated
  USING ((SELECT auth.uid()) = user_id);

ALTER POLICY "Users can insert own projects" ON projects
  TO authenticated
  WITH CHECK ((SELECT auth.uid()) = user_id);

ALTER POLICY "Users can update own projects" ON projects
  TO authenticated
  USING ((SELECT auth.uid()) = user_id);

ALTER POLICY "Users can delete own projects" ON projects
  TO authenticated
  USING ((SELECT auth.uid()) = user_id);

ALTER POLICY "Service role full access projects" ON projects
  TO service_role
  USING (true)
  WITH CHECK (true);

-- ---- MODULS ----

ALTER POLICY "Users can view own moduls" ON moduls
  TO authenticated
  USING ((SELECT auth.uid()) = user_id);

ALTER POLICY "Users can create own moduls" ON moduls
  TO authenticated
  WITH CHECK ((SELECT auth.uid()) = user_id);

ALTER POLICY "Users can update own moduls" ON moduls
  TO authenticated
  USING ((SELECT auth.uid()) = user_id);

ALTER POLICY "Users can delete own moduls" ON moduls
  TO authenticated
  USING ((SELECT auth.uid()) = user_id);

ALTER POLICY "Service role full access moduls" ON moduls
  TO service_role
  USING (true)
  WITH CHECK (true);

-- ---- TUS_UPLOADS ----

ALTER POLICY "Users can view own uploads" ON tus_uploads
  TO authenticated
  USING ((SELECT auth.uid()) = user_id);

ALTER POLICY "Users can insert own uploads" ON tus_uploads
  TO authenticated
  WITH CHECK ((SELECT auth.uid()) = user_id);

ALTER POLICY "Users can update own uploads" ON tus_uploads
  TO authenticated
  USING ((SELECT auth.uid()) = user_id);

ALTER POLICY "Users can delete own uploads" ON tus_uploads
  TO authenticated
  USING ((SELECT auth.uid()) = user_id);

ALTER POLICY "Service role full access uploads" ON tus_uploads
  TO service_role
  USING (true)
  WITH CHECK (true);

-- ---- TUS_MODUL_UPLOADS ----

ALTER POLICY "Users can view own tus modul uploads" ON tus_modul_uploads
  TO authenticated
  USING ((SELECT auth.uid()) = user_id);

ALTER POLICY "Users can create own tus modul uploads" ON tus_modul_uploads
  TO authenticated
  WITH CHECK ((SELECT auth.uid()) = user_id);

ALTER POLICY "Users can update own tus modul uploads" ON tus_modul_uploads
  TO authenticated
  USING ((SELECT auth.uid()) = user_id);

ALTER POLICY "Users can delete own modul uploads" ON tus_modul_uploads
  TO authenticated
  USING ((SELECT auth.uid()) = user_id);

ALTER POLICY "Service role full access tus modul uploads" ON tus_modul_uploads
  TO service_role
  USING (true)
  WITH CHECK (true);
