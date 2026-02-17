-- ================================================
-- RLS Policy Migration: user_profiles + Admin-Only Tables
-- Fix: InitPlan caching (RLS-01), service role scoping (RLS-02), user role scoping (RLS-03)
-- Tables: user_profiles, roles, permissions, casbin_rule, role_permissions
-- Applied via Supabase MCP apply_migration on 2026-02-17
-- ================================================

-- ---- USER_PROFILES (Special Case) ----

-- Fix user CRUD policies (RLS-01 + RLS-03)
-- NOTE: user_profiles uses 'id' not 'user_id' as the ownership column
ALTER POLICY "Users can view own profile" ON user_profiles
  TO authenticated
  USING ((SELECT auth.uid()) = id);

ALTER POLICY "Users can insert own profile" ON user_profiles
  TO authenticated
  WITH CHECK ((SELECT auth.uid()) = id);

ALTER POLICY "Users can update own profile" ON user_profiles
  TO authenticated
  USING ((SELECT auth.uid()) = id);

-- Fix service role FOR ALL policy (RLS-02)
ALTER POLICY "Service role full access" ON user_profiles
  TO service_role
  USING (true)
  WITH CHECK (true);

-- Remove duplicate service role DELETE policy (redundant with FOR ALL above)
DROP POLICY IF EXISTS "Service role can delete profiles" ON user_profiles;

-- ---- ADMIN-ONLY TABLES ----

ALTER POLICY "Service role full access roles" ON roles
  TO service_role
  USING (true)
  WITH CHECK (true);

ALTER POLICY "Service role full access permissions" ON permissions
  TO service_role
  USING (true)
  WITH CHECK (true);

ALTER POLICY "Service role full access casbin_rule" ON casbin_rule
  TO service_role
  USING (true)
  WITH CHECK (true);

ALTER POLICY "Service role full access role_permissions" ON role_permissions
  TO service_role
  USING (true)
  WITH CHECK (true);
