-- 系统通知模块
-- 1. sys_notice 存储通知主数据
-- 2. sys_notice_target 存储目标规则（全体、用户、部门、角色）
-- 3. sys_notice_recipient 存储发布时展开后的用户接收快照

CREATE TABLE IF NOT EXISTS `sys_notice` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `title` varchar(100) NOT NULL COMMENT '通知标题',
  `content` longtext NOT NULL COMMENT '通知内容',
  `type` tinyint(4) NOT NULL COMMENT '通知类型',
  `level` varchar(10) NOT NULL COMMENT '通知等级',
  `publisher_id` int(11) unsigned DEFAULT '0' COMMENT '发布人ID',
  `publish_status` tinyint(4) NOT NULL DEFAULT '0' COMMENT '发布状态: 0未发布 1已发布 -1已撤回',
  `publish_time` datetime DEFAULT NULL COMMENT '发布时间',
  `revoke_time` datetime DEFAULT NULL COMMENT '撤回时间',
  `tenant_id` int(11) unsigned DEFAULT '0' COMMENT '租户ID',
  `created_by` int(11) unsigned DEFAULT '0' COMMENT '创建人ID',
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  `deleted_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`) USING BTREE,
  KEY `idx_notice_tenant_status_time` (`tenant_id`,`publish_status`,`publish_time`) USING BTREE,
  KEY `idx_notice_deleted_at` (`deleted_at`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci ROW_FORMAT=DYNAMIC COMMENT='系统通知主表';

CREATE TABLE IF NOT EXISTS `sys_notice_target` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `notice_id` int(11) unsigned NOT NULL COMMENT '通知ID',
  `target_type` tinyint(4) NOT NULL COMMENT '目标类型: 1全体 2用户 3部门 4角色',
  `target_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '目标ID',
  `include_children` tinyint(1) NOT NULL DEFAULT '0' COMMENT '部门场景是否包含下级',
  `tenant_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '租户ID',
  `created_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE KEY `uk_notice_target` (`notice_id`,`target_type`,`target_id`,`include_children`) USING BTREE,
  KEY `idx_notice_target_lookup` (`tenant_id`,`target_type`,`target_id`) USING BTREE,
  KEY `idx_notice_target_notice_id` (`notice_id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci ROW_FORMAT=DYNAMIC COMMENT='系统通知目标规则表';

CREATE TABLE IF NOT EXISTS `sys_notice_recipient` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `notice_id` int(11) unsigned NOT NULL COMMENT '通知ID',
  `user_id` int(11) unsigned NOT NULL COMMENT '接收人ID',
  `read_status` tinyint(4) NOT NULL DEFAULT '0' COMMENT '读取状态: 0未读 1已读',
  `read_time` datetime DEFAULT NULL COMMENT '阅读时间',
  `tenant_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '租户ID',
  `publish_time` datetime DEFAULT NULL COMMENT '发布时间快照',
  `created_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE KEY `uk_notice_user` (`notice_id`,`user_id`) USING BTREE,
  KEY `idx_notice_recipient_user` (`tenant_id`,`user_id`,`read_status`,`publish_time`) USING BTREE,
  KEY `idx_notice_recipient_notice_id` (`notice_id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci ROW_FORMAT=DYNAMIC COMMENT='系统通知接收人表';
