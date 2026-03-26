-- sys_customer index optimization
-- Date: 2026-03-26
--
-- Design notes
-- 1. The current application always applies tenant filtering in normal list/detail queries.
-- 2. GORM soft delete adds `deleted_at IS NULL` to model queries.
-- 3. Current `num` / `mobile` / `name` conditions use LIKE '%keyword%', which cannot
--    reliably use normal BTREE indexes. Those are left out of the core migration.
-- 4. This migration focuses on the filters that are actually index-friendly today.

START TRANSACTION;

-- Remove tenant-blind single-column indexes that are weaker than the tenant-aware composites below.
ALTER TABLE `sys_customer`
  DROP INDEX `idx_user_id`,
  DROP INDEX `idx_channel_id`,
  DROP INDEX `idx_department_id`,
  DROP INDEX `idx_status`,
  DROP INDEX `idx_md5_mobile`;

-- Core composite indexes for current query patterns.
ALTER TABLE `sys_customer`
  ADD INDEX `idx_tenant_user_deleted` (`tenant_id`, `user_id`, `deleted_at`),
  ADD INDEX `idx_tenant_channel_deleted` (`tenant_id`, `channel_id`, `deleted_at`),
  ADD INDEX `idx_tenant_dept_deleted` (`tenant_id`, `dept_id`, `deleted_at`),
  ADD INDEX `idx_tenant_status_deleted` (`tenant_id`, `status`, `deleted_at`),
  ADD INDEX `idx_tenant_intention_deleted` (`tenant_id`, `intention`, `deleted_at`),
  ADD INDEX `idx_tenant_reassign_deleted` (`tenant_id`, `is_reassign`, `deleted_at`),
  ADD INDEX `idx_tenant_star_status_deleted` (`tenant_id`, `star_status`, `deleted_at`),
  ADD INDEX `idx_tenant_md5_mobile_deleted` (`tenant_id`, `md5_mobile`, `deleted_at`);

COMMIT;

-- Keep `idx_mobile` for now.
-- Reason:
-- 1. The current code uses LIKE '%xxx%', so this index is not a reliable accelerator.
-- 2. It may still be useful if you later switch some queries to `mobile = ?` or `mobile LIKE '138%'`.
--
-- Optional scene indexes:
-- Add these only if EXPLAIN shows slow scans on public/exchange/locked lists inside a single tenant.
--
-- ALTER TABLE `sys_customer`
--   ADD INDEX `idx_tenant_public_deleted` (`tenant_id`, `is_public`, `deleted_at`),
--   ADD INDEX `idx_tenant_exchange_deleted` (`tenant_id`, `is_exchange`, `deleted_at`),
--   ADD INDEX `idx_tenant_lock_deleted` (`tenant_id`, `is_lock`, `deleted_at`);
--
-- Optional if you later change query style:
-- 1. If `num` becomes exact or prefix search, add:
--    ALTER TABLE `sys_customer`
--      ADD INDEX `idx_tenant_num_deleted` (`tenant_id`, `num`, `deleted_at`);
-- 2. If `mobile` becomes exact lookup, prefer:
--    ALTER TABLE `sys_customer`
--      ADD INDEX `idx_tenant_mobile_deleted` (`tenant_id`, `mobile`, `deleted_at`);
-- 3. If you keep privacy lookup, prefer querying `md5_mobile = ?` instead of `mobile LIKE '%...%'`.
