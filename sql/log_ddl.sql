-- domain_monitor_history definition

CREATE TABLE `domain_monitor_history` (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `created_at` datetime NOT NULL, `updated_at` datetime NOT NULL, `monitor_id` integer NOT NULL, `check_time` datetime NOT NULL, `is_success` bool NOT NULL, `check_duration` integer NULL, `status_message` text NULL, `http_status` integer NULL, `ip_resolved` json NULL, `error_log` text NULL, `ssl_expiry_date` datetime NULL, `ssl_common_name` text NULL, `ssl_issuer_name` text NULL, `ssl_signature_alg` text NULL, `ssl_san_list` text NULL, `ssl_certificate_chain` text NULL);

CREATE INDEX `domainmonitorhistory_check_time` ON `domain_monitor_history` (`check_time` DESC);
CREATE INDEX `domainmonitorhistory_monitor_id_check_time` ON `domain_monitor_history` (`monitor_id`, `check_time`);


-- http_log definition

CREATE TABLE "http_log" (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `hostname` text NOT NULL, `method` text NOT NULL, `url` text NOT NULL, `request_headers` json NULL, `request_body` json NULL, `status_code` integer NOT NULL, `response_headers` json NULL, `response_body` json NULL, `elapse_time` integer NOT NULL, `is_error` bool NULL DEFAULT (false), `created_at` datetime NOT NULL, `go_version` text NULL DEFAULT ('go1.26rc1'), `plugin_version` text NULL DEFAULT ('1.0.0'));

CREATE INDEX `httplog_hostname` ON `http_log` (`hostname`);
CREATE INDEX `httplog_method` ON `http_log` (`method`);
CREATE INDEX `httplog_url` ON `http_log` (`url`);
CREATE INDEX `httplog_status_code` ON `http_log` (`status_code`);
CREATE INDEX `httplog_created_at` ON `http_log` (`created_at` DESC);


-- operation_log definition

CREATE TABLE "operation_log" (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `created_at` datetime NOT NULL, `op_type` text NOT NULL, `obj_id` integer NOT NULL, `obj_type` text NOT NULL, `old_value` text NULL, `new_value` text NULL, `status` text NOT NULL, `message` text NOT NULL);

CREATE INDEX `operationlog_op_type` ON `operation_log` (`op_type`);
CREATE INDEX `operationlog_obj_id` ON `operation_log` (`obj_id`);
CREATE INDEX `operationlog_obj_type` ON `operation_log` (`obj_type`);
CREATE INDEX `operationlog_created_at` ON `operation_log` (`created_at` DESC);


-- resty_log definition

CREATE TABLE "resty_log" (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `hostname` text NOT NULL, `method` text NOT NULL, `url` text NOT NULL, `request_headers` json NULL, `request_body` json NULL, `status_code` integer NOT NULL, `response_headers` json NULL, `response_body` json NULL, `elapse_time` integer NOT NULL, `is_error` bool NULL DEFAULT (false), `created_at` datetime NOT NULL, `go_version` text NULL DEFAULT ('go1.26rc1'), `plugin_version` text NULL DEFAULT ('1.0.10'), `resty_version` text NULL DEFAULT ('3.0.0-beta.5'));

CREATE INDEX `restylog_hostname` ON `resty_log` (`hostname`);
CREATE INDEX `restylog_method` ON `resty_log` (`method`);
CREATE INDEX `restylog_url` ON `resty_log` (`url`);
CREATE INDEX `restylog_status_code` ON `resty_log` (`status_code`);
CREATE INDEX `restylog_created_at` ON `resty_log` (`created_at` DESC);


-- webhook_history definition

CREATE TABLE `webhook_history` (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `created_at` datetime NOT NULL, `updated_at` datetime NOT NULL, `webhook_id` integer NOT NULL, `subscription_id` integer NULL, `url` text NOT NULL, `title` text NULL, `content` text NULL, `status_code` integer NULL, `response` text NULL, `success` bool NOT NULL DEFAULT (false), `error_message` text NULL);

CREATE INDEX `webhookhistory_webhook_id` ON `webhook_history` (`webhook_id`);
CREATE INDEX `webhookhistory_subscription_id` ON `webhook_history` (`subscription_id`);
CREATE INDEX `webhookhistory_success` ON `webhook_history` (`success`);
CREATE INDEX `webhookhistory_created_at` ON `webhook_history` (`created_at` DESC);
CREATE INDEX `webhookhistory_webhook_id_created_at` ON `webhook_history` (`webhook_id`, `created_at`);
CREATE INDEX `webhookhistory_success_created_at` ON `webhook_history` (`success`, `created_at`);