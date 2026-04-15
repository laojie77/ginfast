ALTER TABLE `sys_customer_import_batches`
  ADD COLUMN `start_row` int NOT NULL DEFAULT 2 COMMENT '导入起始行号' AFTER `channel_name`,
  ADD COLUMN `processed_count` int NOT NULL DEFAULT 0 COMMENT '已处理行数' AFTER `total_count`,
  ADD COLUMN `progress` int NOT NULL DEFAULT 0 COMMENT '当前进度百分比' AFTER `failed_count`,
  ADD COLUMN `resume_row` int NOT NULL DEFAULT 0 COMMENT '建议继续导入的行号' AFTER `progress`,
  ADD COLUMN `failure_preview` longtext NOT NULL COMMENT '失败预览JSON' AFTER `error_message`,
  ADD COLUMN `started_at` datetime(3) DEFAULT NULL COMMENT '导入开始时间' AFTER `failure_preview`,
  ADD COLUMN `finished_at` datetime(3) DEFAULT NULL COMMENT '导入结束时间' AFTER `started_at`;
